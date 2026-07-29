package kfkit

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Reader struct {
	reader *kafka.Reader

	ctx    context.Context
	cancel context.CancelFunc
}

func (r *Reader) Handle(handle func(msg kafka.Message) error) {
	for {
		select {
		case <-r.ctx.Done():
			slog.Warn("kafka reader stopped", slog.Any("error", r.ctx.Err()))
			return
		default:
			msg, err := r.reader.ReadMessage(r.ctx)

			// io.EOF means reader closed.
			// io.ErrClosedPipe means committing messages on the reader,
			// kafka will refire the messages on uncommitted messages, ignore
			if err == io.EOF || err == io.ErrClosedPipe {
				slog.Warn("kafka reader closed", slog.Any("error", err))
				return
			}
			if err != nil {
				slog.Error("failed to read kafka message", slog.Any("error", err))
				continue
			}

			start := time.Now()

			metricSubDelay.WithLabelValues(msg.Topic).Observe(float64(time.Since(msg.Time).Milliseconds()))

			result := "success"
			if err = handle(msg); err != nil {
				result = "fail"
			}
			metricResult.WithLabelValues(msg.Topic, "sub", result).Inc()

			metricReqDuration.WithLabelValues(msg.Topic, "sub").Observe(float64(time.Since(start).Milliseconds()))
		}
	}
}

func (r *Reader) Close() error {
	r.cancel()
	return r.reader.Close()
}

func DefaultReaderConfig(brokers []string, groupId, topic string) kafka.ReaderConfig {
	return kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupId,
		MinBytes:       10e3, // 10K
		MaxBytes:       10e6, // 10MB
		MaxWait:        time.Second,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	}
}

func NewReader(config kafka.ReaderConfig) *Reader {
	ctx, cancel := context.WithCancel(context.Background())

	return &Reader{
		reader: kafka.NewReader(config),

		ctx:    ctx,
		cancel: cancel,
	}
}
