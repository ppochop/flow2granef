package input

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/prometheus/client_golang/prometheus"
)

type KafkaInput struct {
	consumer     *kafka.Consumer
	msgsConsumed prometheus.Counter
}

//var BootstrapServers string
//var GroupID string
//var Topic string

func init() {
	//flag.StringVar(&BootstrapServers, "kafka-bootstrap-servers", "localhost", "Kafka bootstrap servers. If more than one, separate with commas")
	//flag.StringVar(&GroupID, "kafka-groupid", "", "Kafka Consumer group id")
	//flag.StringVar(&Topic, "kafka-topic", "", "Kafka topic to consume")
	RegisterInput("kafka", InitKafkaConsumer)
}

type KafkaConfig struct {
	BootstrapServers string `toml:"bootstrap-servers"` // Kafka bootstrap servers. If more than one, separate with commas
	GroupId          string `toml:"group-id"`          // Kafka Consumer group id
	Topic            string `toml:"topic"`             // Kafka topic to consume
}

func InitKafkaConsumer(config InputConfig, stats InputStats) (Input, error) {
	kC := KafkaConfig{
		BootstrapServers: config["bootstrap-servers"].(string),
		GroupId:          config["group-id"].(string),
		Topic:            config["topic"].(string),
	}
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":    kC.BootstrapServers,
		"group.id":             kC.GroupId,
		"max.poll.interval.ms": 80000000,
		"auto.offset.reset":    "latest",
	})
	if err != nil {
		return nil, err
	}
	err = c.Subscribe(kC.Topic, nil)
	if err != nil {
		return nil, err
	}
	return &KafkaInput{
		consumer:     c,
		msgsConsumed: stats.MsgsConsumed,
	}, nil
}

func (k *KafkaInput) NextEntry() ([]byte, error) {

	msg, err := k.consumer.ReadMessage(time.Second)
	if err == nil {
		k.msgsConsumed.Inc()
		return msg.Value, nil
	} else if !err.(kafka.Error).IsTimeout() {
		return nil, err
	}
	return nil, nil
}
