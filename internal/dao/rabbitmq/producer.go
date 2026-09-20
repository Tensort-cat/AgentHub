package rabbitmq

import (
    "encoding/json"
    "errors"
    "fmt"
    "time"

    amqp "github.com/rabbitmq/amqp091-go"
)

func PublishWorkflowRun(task WorkflowRunTask) error {
    if task.TaskID == "" {
        task.TaskID = fmt.Sprintf("wf-%d-%d", time.Now().UnixNano(), task.WorkflowID)
    }
    if task.CreatedAt.IsZero() {
        task.CreatedAt = time.Now().UTC()
    }

    body, err := json.Marshal(task)
    if err != nil {
        return err
    }

    return publishToQueue(QueueWorkflowRun, task.TaskID, body)
}

func PublishWorkflowResult(result WorkflowRunResult) error {
    if result.TaskID == "" {
        result.TaskID = fmt.Sprintf("wf-result-%d-%d", time.Now().UnixNano(), result.WorkflowID)
    }
    if result.FinishedAt.IsZero() {
        result.FinishedAt = time.Now().UTC()
    }

    body, err := json.Marshal(result)
    if err != nil {
        return err
    }

    return publishToQueue(QueueWorkflowResult, result.TaskID, body)
}

func publishToQueue(queueName, messageID string, body []byte) error {
    if RabbitMQConn == nil {
        return errors.New("rabbitmq connection is nil")
    }

    ch, err := RabbitMQConn.Channel()
    if err != nil {
        return err
    }
    defer ch.Close()

    return ch.Publish(
        "",
        queueName,
        false,
        false,
        amqp.Publishing{
            DeliveryMode: amqp.Persistent,
            ContentType:  "application/json",
            Timestamp:    time.Now(),
            MessageId:    messageID,
            Body:         body,
        },
    )
}
