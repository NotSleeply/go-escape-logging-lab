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

// Request 是三个日志库共用的结构化日志数据。
type Request struct {
	ID      string
	Path    string
	Status  int
	Latency time.Duration
}

// SampleRequest 返回固定的示例请求，便于观察三种日志输出形式。
func SampleRequest() Request {
	return Request{
		ID:      "req-9f5d20",
		Path:    "/v1/orders/42",
		Status:  200,
		Latency: 3 * time.Millisecond,
	}
}

// NewZapLogger 创建输出 JSON 的 zap 日志器。
func NewZapLogger(out io.Writer) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(encoder, zapcore.AddSync(out), zap.InfoLevel)
	return zap.New(core)
}

// NewSlogLogger 创建输出 JSON 的 slog 日志器。
func NewSlogLogger(out io.Writer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// NewStdLogger 创建不带前缀和时间标记的标准库日志器。
func NewStdLogger(out io.Writer) *log.Logger {
	return log.New(out, "", 0)
}

// main 使用相同请求数据演示 log、slog 与 zap 的原生调用方式。
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
