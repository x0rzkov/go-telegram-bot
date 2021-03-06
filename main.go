package main

import (
	"bytes"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/welly87/go-telegram-bot/avro"

	//avro "github.com/welly87/go-telegram-bot"
	"log"
	"time"
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func main() {
	bot, err := tgbotapi.NewBotAPI("997919151:AAGd-cGPlqq42GKDKD8YGjnpSaD40xjVF18")
	if err != nil {
		log.Panic(err)
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "34.87.113.63"})
	if err != nil {
		panic(err)
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	defer p.Close()

	// Produce messages to topic (asynchronously)
	topic := "telegram"

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		//update.Message.

		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID

		// bot.Send(msg)

		msg := avro.NewTelegramMessage()
		msg.Message = update.Message.Text
		msg.ChatId = string(update.Message.Chat.ID)
		msg.MessageId = string(update.Message.MessageID)
		msg.Username = update.Message.From.UserName
		msg.Viewtime = makeTimestamp()

		var buf bytes.Buffer
		msg.Serialize(&buf)

		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          buf.Bytes(),
		}, nil)
	}
}
