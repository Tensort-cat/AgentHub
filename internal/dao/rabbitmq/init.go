package rabbitmq

import (
	"AgentHub/internal/config"
	"AgentHub/pkg/zlog"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitMQConn *amqp.Connection

func InitRabbitmq() error {
	cfg := config.Cfg.RabbitmqConfig

	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	RabbitMQConn = conn

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	defer ch.Close()

	if err := CreateDocQueues(ch); err != nil {
		return err
	}

	// if err := CreateRunQueue(ch); err != nil {
	// 	return err
	// }

	return nil
}

func Close() error {
	return RabbitMQConn.Close()
}

func CreateDocQueues(ch *amqp.Channel) error {

	queues := []string{
		DocParseQueue,
		DocIndexQueue,
	}

	for _, queue := range queues {

		_, err := ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			amqp.Table{
				amqp.QueueTypeArg: amqp.QueueTypeQuorum,
			},
		)

		if err != nil {
			zlog.Error(err.Error())
			return err
		}
	}

	return nil
}

func CreateRunQueue(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(
		"run_queue",
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		zlog.Error(err.Error())
		return err
	}

	return nil
}
