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

	if err := declareQueue(RabbitMQConn, QueueWorkflowRun); err != nil {
		return err
	}
	if err := declareQueue(RabbitMQConn, QueueWorkflowResult); err != nil {
		return err
	}

	return nil
}

func Close() error {
	if RabbitMQConn == nil {
		return nil
	}
	return RabbitMQConn.Close()
}

func declareQueue(conn *amqp.Connection, queueName string) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		amqp.Table{
			amqp.QueueTypeArg: amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		zlog.Error("declare queue failed: " + queueName + ", err=" + err.Error())
		return err
	}

	return nil
}
