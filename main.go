package main

import (
	"encoding/json"
	"fmt"
	"kafka-consumer/consumer"
	"kafka-consumer/db"
	"kafka-consumer/processor"
	"os"
	"os/signal"
)

func main() {

	kafkaBroker := []string{"x:x"}
	consumeTopics := []string{"test"}

	kafkaConsumer := consumer.NewKafkaConsumer(kafkaBroker, "DT01", consumeTopics)
	defer kafkaConsumer.Close()

	s := db.InitDB("x:x")
	defer s.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	dataProcessor := processor.NewProcessor(2)
	dataProcessor.Run()

	for {
		select {
		case msg, ok := <-kafkaConsumer.Messages():
			if ok {
				msgData := processor.MonitorData{}
				err := json.Unmarshal(msg.Value, &msgData)
				if err != nil {
					fmt.Println("Invalid Data")
					continue
				}
				dataProcessor.AddData(&msgData)

				fmt.Fprintf(os.Stdout, "接收Kafka信息：主题-%s/分区-%d/偏移-%d\t消息-Key:%s\tValue:%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
				kafkaConsumer.MarkOffset(msg, "") // mark message as processed
			}
		case <-signals:
			fmt.Println("Get Signal, Wait For Processor")
			dataProcessor.Wait()
			fmt.Println("Processor Graceful Down, Bye")

			return
		}
	}
}
