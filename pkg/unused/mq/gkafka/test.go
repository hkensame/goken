package gkafka

// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"os"
// 	"os/signal"

// 	"github.com/Shopify/sarama"
// )

// var (
// 	brokers = []string{"localhost:9092"}
// 	groupID = "example-group"
// 	topic   = "example-topic"
// )

// func main() {
// 	// 创建配置
// 	config := sarama.NewConfig()
// 	config.Version = sarama.V2_5_0_0 // 设置 Kafka 版本（要与你的 broker 匹配）
// 	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
// 	config.Consumer.Offsets.Initial = sarama.OffsetOldest // 第一次从头消费

// 	// 创建 ConsumerGroup 实例
// 	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
// 	if err != nil {
// 		log.Fatalf("Error creating consumer group: %v", err)
// 	}
// 	defer consumerGroup.Close()

// 	// 启动消费
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	go func() {
// 		for err := range consumerGroup.Errors() {
// 			log.Printf("Consumer group error: %v", err)
// 		}
// 	}()

// 	handler := &ExampleConsumerGroupHandler{}

// 	go func() {
// 		for {
// 			err := consumerGroup.Consume(ctx, []string{topic}, handler)
// 			if err != nil {
// 				log.Printf("Error from Consume: %v", err)
// 			}
// 			// 必须处理退出信号，否则 rebalance 会崩
// 			if ctx.Err() != nil {
// 				return
// 			}
// 		}
// 	}()

// 	// 等待退出信号
// 	sigterm := make(chan os.Signal, 1)
// 	signal.Notify(sigterm, os.Interrupt)
// 	<-sigterm
// 	log.Println("Shutting down...")
// }

// // 实现 ConsumerGroupHandler 接口
// type ExampleConsumerGroupHandler struct{}

// func (h *ExampleConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
// 	log.Println("Consumer group setup")
// 	return nil
// }

// func (h *ExampleConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
// 	log.Println("Consumer group cleanup")
// 	return nil
// }

// func (h *ExampleConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
// 	// 消费主循环
// 	for msg := range claim.Messages() {
// 		fmt.Printf("Message topic:%q partition:%d offset:%d key:%s value:%s\n",
// 			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
// 		// 手动提交当前 offset
// 		session.MarkMessage(msg, "")
// 	}
// 	return nil
// }
