package rabbitmq

import (
	"AgentHub/pkg/zlog"
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"go.uber.org/zap"
)

type WorkflowRunHandler func(ctx context.Context, task WorkflowRunTask) error
type WorkflowResultHandler func(ctx context.Context, result WorkflowRunResult) error

// 启动消费者
func StartWorkflowConsumer(ctx context.Context, concurrency int, handler WorkflowRunHandler) error {
	if RabbitMQConn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if handler == nil {
		return errors.New("workflow run handler is nil")
	}

	for i := 0; i < concurrency; i++ {
		go consumeWorkflowRunQueue(ctx, i, handler)
	}

	return nil
}

func StartWorkflowResultConsumer(ctx context.Context, handler WorkflowResultHandler) error {
	if RabbitMQConn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	if handler == nil {
		return errors.New("workflow result handler is nil")
	}

	go consumeWorkflowResultQueue(ctx, handler)
	return nil
}

// 消费 Run 任务
func consumeWorkflowRunQueue(ctx context.Context, index int, handler WorkflowRunHandler) {
	ch, err := RabbitMQConn.Channel()
	if err != nil {
		zlog.Error("create RabbitMQ channel failed: " + err.Error())
		return
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		zlog.Error("set QoS failed: " + err.Error())
		return
	}

	deliveries, err := ch.Consume(
		QueueWorkflowRun,
		"workflow_consumer_"+strconv.Itoa(index),
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zlog.Error("start consumer failed: " + err.Error())
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}

			var task WorkflowRunTask
			if err := json.Unmarshal(delivery.Body, &task); err != nil {
				zlog.Error("unmarshal task failed: "+err.Error(), zap.ByteString("body", delivery.Body))
				_ = delivery.Nack(false, false)
				continue
			}

			if err := handler(ctx, task); err != nil {
				zlog.Error("handle workflow task failed: "+err.Error(), zap.String("task_id", task.TaskID))
				_ = delivery.Nack(false, true)
				continue
			}

			if err := delivery.Ack(false); err != nil {
				zlog.Error("ACK failed: "+err.Error(), zap.String("task_id", task.TaskID))
			}
		}
	}
}

func consumeWorkflowResultQueue(ctx context.Context, handler WorkflowResultHandler) {
	ch, err := RabbitMQConn.Channel()
	if err != nil {
		zlog.Error("create RabbitMQ result channel failed: " + err.Error())
		return
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		zlog.Error("set result consumer QoS failed: " + err.Error())
		return
	}

	deliveries, err := ch.Consume(
		QueueWorkflowResult,
		"workflow_result_consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		zlog.Error("start result consumer failed: " + err.Error())
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				return
			}

			var result WorkflowRunResult
			if err := json.Unmarshal(delivery.Body, &result); err != nil {
				zlog.Error("unmarshal workflow result failed: "+err.Error(), zap.ByteString("body", delivery.Body))
				_ = delivery.Nack(false, false)
				continue
			}

			if err := handler(ctx, result); err != nil {
				zlog.Error("handle workflow result failed: "+err.Error(), zap.String("task_id", result.TaskID))
				_ = delivery.Nack(false, true)
				continue
			}

			if err := delivery.Ack(false); err != nil {
				zlog.Error("result ACK failed: "+err.Error(), zap.String("task_id", result.TaskID))
			}
		}
	}
}
