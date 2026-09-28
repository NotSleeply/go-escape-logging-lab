package main

import (
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Request struct {
	ID      string
	Path    string
	Status  int
	Latency time.Duration
}

func SampleRequest() Request {
	return Request{
		ID:      "req-9f5d20",
		Path:    "/v1/orders/42",
		Status:  200,
		Latency: 3 * time.Millisecond,
	}
}

func NewZapLogger(out io.Writer) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(out), zap.InfoLevel)
	return zap.New(core)
}

func NewSlogLogger(out io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func NewStdLogger(out io.Writer) *log.Logger {
	return log.New(out, "", 0)
}

func main() {
	request := SampleRequest()

	standard := NewStdLogger(os.Stdout)
	standard.Printf("log: request completed request_id=%s path=%s status=%d latency=%s",
		request.ID, request.Path, request.Status, request.Latency)

	structured := NewSlogLogger(os.Stdout)
	structured.LogAttrs(nil, slog.LevelInfo, "slog: request completed",
		slog.String("request_id", request.ID),
		slog.String("path", request.Path),
		slog.Int("status", request.Status),
		slog.Duration("latency", request.Latency),
	)

	fast := NewZapLogger(os.Stdout)
	fast.Info("zap: request completed",
		zap.String("request_id", request.ID),
		zap.String("path", request.Path),
		zap.Int("status", request.Status),
		zap.Duration("latency", request.Latency),
	)
	_ = fast.Sync()
}
