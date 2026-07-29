package kfkit

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func (p *Producer) Send(ctx context.Context, topic string, key string, payload []byte) (err error) {
	start := time.Now()

	defer func() {
		level := slog.LevelInfo
		attrs := []slog.Attr{
			slog.String("cost", time.Since(start).String()),
			slog.String("topic", topic),
			slog.String("key", key),
			slog.String("payload", string(payload)),
		}
		if err != nil {
			level = slog.LevelError
			attrs = append(attrs, slog.Any("err", err))
		}
		slog.LogAttrs(ctx, level, "sendkafka", attrs...)
	}()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	})

	result := "success"
	if err != nil {
		result = "fail"
	}
	metricResult.WithLabelValues(topic, pub, result).Inc()
	metricReqDuration.WithLabelValues(topic, pub).Observe(float64(time.Since(start).Milliseconds()))

	return
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func DefaultWriterConfig(brokers []string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		BatchSize:    1000,
		BatchBytes:   5242880,
		BatchTimeout: 20 * time.Millisecond,
		Async:        true,
	}
}

func NewProducer(w *kafka.Writer) *Producer {
	return &Producer{
		writer: w,
	}
}
