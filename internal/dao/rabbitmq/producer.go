package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DocParseQueue = "doc_parse_queue"
	DocIndexQueue = "doc_index_queue"
)

func PublishParseTask(
	ctx context.Context,
	msg DocumentMessage,
) error {
	return publish(ctx, DocParseQueue, msg)
}

func PublishIndexTask(
	ctx context.Context,
	msg DocumentMessage,
) error {
	return publish(ctx, DocIndexQueue, msg)
}

func publish(
	ctx context.Context,
	queue string,
	msg DocumentMessage,
) error {

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal rabbitmq message: %w", err)
	}

	ch, err := RabbitMQConn.Channel()
	if err != nil {
		return fmt.Errorf("create rabbitmq channel: %w", err)
	}
	defer ch.Close()

	err = ch.PublishWithContext(
		ctx,
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    msg.TaskID,
			Timestamp:    time.Now(),

			Body: body,
		},
	)

	if err != nil {
		return fmt.Errorf("publish rabbitmq message: %w", err)
	}

	return nil
}
