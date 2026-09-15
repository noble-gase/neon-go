package kfkit

import (
	"context"
	"log/slog"
	"time"

	"github.com/noble-gase/neon/helper"
	"github.com/noble-gase/xe/worker"
	"github.com/segmentio/kafka-go"
)

type Handler = func(ctx context.Context, msg kafka.Message) error

// Consumer Kafka消费者
type Consumer struct {
	Group string
	Topic []Topic
}

type Topic struct {
	Name    string
	Handler Handler
}

// InitConsumer 同步消费
func InitConsumer(brokers []string, consumers ...Consumer) {
	for _, c := range consumers {
		for _, t := range c.Topic {
			conf := DefaultReaderConfig(brokers, c.Group, t.Name)
			handle := HandleMessage(c.Group, t.Handler)

			reader := NewReader(conf)
			go reader.Handle(handle)
		}
	}
}

func HandleMessage(group string, handler Handler) func(msg kafka.Message) error {
	return func(msg kafka.Message) (err error) {
		ctx, cancel := context.WithTimeout(helper.CtxWithTraceID(context.Background()), 10*time.Second)
		defer cancel()

		now := time.Now()
		delay := time.Since(msg.Time)

		defer func() {
			level := slog.LevelInfo
			attrs := []slog.Attr{
				slog.String("delay", delay.String()),
				slog.String("cost", time.Since(now).String()),
				slog.String("topic", msg.Topic),
				slog.String("group", group),
				slog.String("key", string(msg.Key)),
				slog.Any("headers", msg.Headers),
				slog.String("payload", string(msg.Value)),
				slog.Int("partition", msg.Partition),
			}
			if err != nil {
				level = slog.LevelError
				attrs = append(attrs, slog.Any("err", err))
			}
			slog.LogAttrs(ctx, level, "recvkafka", attrs...)
		}()

		err = handler(ctx, msg)
		return
	}
}

// InitAsyncConsumer 异步并发消费
func InitAsyncConsumer(pool worker.Pool, brokers []string, consumers ...Consumer) {
	for _, c := range consumers {
		for _, t := range c.Topic {
			conf := DefaultReaderConfig(brokers, c.Group, t.Name)
			handle := AsyncHandleMessage(pool, c.Group, t.Handler)

			reader := NewReader(conf)
			go reader.Handle(handle)
		}
	}
}

func AsyncHandleMessage(pool worker.Pool, group string, handler Handler) func(msg kafka.Message) error {
	return func(msg kafka.Message) error {
		ctx := helper.CtxWithTraceID(context.Background())

		_ = pool.Go(ctx, func(ctx context.Context) {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			var err error

			now := time.Now()
			delay := time.Since(msg.Time)

			defer func() {
				level := slog.LevelInfo
				attrs := []slog.Attr{
					slog.String("delay", delay.String()),
					slog.String("cost", time.Since(now).String()),
					slog.String("topic", msg.Topic),
					slog.String("group", group),
					slog.String("key", string(msg.Key)),
					slog.Any("headers", msg.Headers),
					slog.String("payload", string(msg.Value)),
					slog.Int("partition", msg.Partition),
				}
				if err != nil {
					level = slog.LevelError
					attrs = append(attrs, slog.Any("err", err))
				}
				slog.LogAttrs(ctx, level, "recvkafka", attrs...)
			}()

			err = handler(ctx, msg)
		})

		return nil
	}
}
