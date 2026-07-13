package producer

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(kafkaBroker string) *Producer {
	// broker := os.Getenv("KAFKA_BROKER")
	// if broker == "" {
	// 	panic(fmt.Errorf("no KAFKA_BROKER"))
	// }

	writer := kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: "messages",
		// как балансируются сообщения по партициям
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		// через сколько перестаём ждать заполнение батча
		BatchTimeout: 10 * time.Millisecond,
	}

	producer := Producer{
		writer: &writer,
	}
	return &producer
}

func (p *Producer) Produce(msg string) {
	// отправка сообщений
	err := p.writer.WriteMessages(context.TODO(), kafka.Message{
		Value: []byte(msg),
	})

	if err != nil {
		panic(err)
	}
}

func (p *Producer) ProduceN(msgs []string) {
	kafkaMsgs := make([]kafka.Message, len(msgs))
	for i, msg := range msgs {
		kafkaMsgs[i] = kafka.Message{
			Value: []byte(msg),
		}
	}

	// отправка сообщений
	err := p.writer.WriteMessages(context.TODO(), kafkaMsgs...)

	if err != nil {
		panic(err)
	}
}

func (p *Producer) Close() {
	p.writer.Close()
}
