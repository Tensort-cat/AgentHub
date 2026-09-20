package rabbitmq

import (
	workflowService "AgentHub/internal/service/workflow"
	"AgentHub/pkg/constant"
	"AgentHub/pkg/zlog"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type WorkflowRunHandler func(ctx context.Context, task WorkflowRunTask) error

func StartWorkflowConsumer(ctx context.Context, concurrency int, handler WorkflowRunHandler) error {
	if RabbitMQConn == nil {
		return errors.New("rabbitmq connection is nil")
	}
	if concurrency < 1 {
		concurrency = 1
	}
	if handler == nil {
		handler = DefaultWorkflowRunHandler
	}

	for i := 0; i < concurrency; i++ {
		go consumeWorkflowRunQueue(ctx, i, handler)
	}

	return nil
}

func DefaultWorkflowRunHandler(ctx context.Context, task WorkflowRunTask) error {
	code, msg, res := workflowService.Run(ctx, task.WorkflowID, task.Input, task.UserID)

	publishResult := WorkflowRunResult{
		TaskID:     task.TaskID,
		WorkflowID: task.WorkflowID,
		UserID:     task.UserID,
		SessionID:  res.SessionID,
		Result:     res.Content,
		Status:     TaskStatusSuccess,
		FinishedAt: time.Now().UTC(),
	}

	if code != constant.Success {
		publishResult.Status = TaskStatusFailed
		publishResult.Error = string(msg)
	}

	if err := PublishWorkflowResult(publishResult); err != nil {
		return err
	}

	if code != constant.Success {
		return errors.New(string(msg))
	}

	return nil
}

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
				_ = delivery.Nack(false, true)
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
