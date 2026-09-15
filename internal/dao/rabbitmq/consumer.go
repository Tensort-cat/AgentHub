package rabbitmq

import (
	"context"

	"AgentHub/pkg/zlog"

	"go.uber.org/zap"
)

func StartIndexConsumer(ctx context.Context) error {

	ch, err := RabbitMQConn.Channel()
	if err != nil {
		return err
	}

	if err := ch.Qos(
		1,
		0,
		false,
	); err != nil {
		ch.Close()
		return err
	}

	msgs, err := ch.Consume(
		DocIndexQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return err
	}

	go func() {
		defer ch.Close()

		for {
			select {
			case <-ctx.Done():
				return

			case delivery, ok := <-msgs:
				if !ok {
					return
				}

				if err := handleIndexMessage(
					ctx,
					delivery,
				); err != nil {

					zlog.Error(
						"index document failed",
						zap.Error(err),
					)

					_ = delivery.Nack(false, true)
					continue
				}

				_ = delivery.Ack(false)
			}
		}
	}()

	return nil
}
