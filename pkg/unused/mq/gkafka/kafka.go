package gkafka

import (
	"github.com/IBM/sarama"
)

func NewSyncProducer(namesrv []string, conf *sarama.Config) sarama.SyncProducer {
	producer, err := sarama.NewSyncProducer(namesrv, conf)
	if err != nil {
		panic(err)
	}
	return producer
}

func NewAsyncProducer(namesrv []string, conf *sarama.Config) sarama.AsyncProducer {
	producer, err := sarama.NewAsyncProducer(namesrv, conf)
	if err != nil {
		panic(err)
	}
	return producer
}

func NewSyncConsumer(namesrv []string, conf *sarama.Config) sarama.Consumer {
	producer, err := sarama.NewSyncProducer(namesrv, conf)
	if err != nil {
		panic(err)
	}
	return producer
}
