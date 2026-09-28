package main

import (
	"fmt"
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

// measurementDiscard 最终将数据写入 io.Discard，但自身不是 io.Discard。
// 这样标准库 log 不会在格式化参数之前直接跳过输出路径。
type measurementDiscard struct {
	target io.Writer
}

func (w measurementDiscard) Write(p []byte) (int, error) {
	return w.target.Write(p)
}

var measurementOutput io.Writer = measurementDiscard{target: io.Discard}

const measurementIterations = 1_000_000

// main 先展示三种原生调用的输出，再测量同一标量字段的运行耗时。
func main() {
	request := SampleRequest()
	showLogExamples(request)

	fmt.Printf("\n运行耗时：每种日志器记录 %d 次，输出写入 io.Discard\n", measurementIterations)
	measureZap(request)
	measureSlog(request)
	measureStdLog(request)
}

// showLogExamples 将三种日志格式各输出一次，便于观察字段编码差异。
func showLogExamples(request Request) {

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

// measureZap 测量 zap 强类型字段的调用耗时。
func measureZap(request Request) {
	logger := NewZapLogger(measurementOutput)
	logger.Info("request completed",
		zap.String("request_id", request.ID),
		zap.String("path", request.Path),
		zap.Int("status", request.Status),
		zap.Duration("latency", request.Latency),
	)

	startedAt := time.Now()
	for range measurementIterations {
		logger.Info("request completed",
			zap.String("request_id", request.ID),
			zap.String("path", request.Path),
			zap.Int("status", request.Status),
			zap.Duration("latency", request.Latency),
		)
	}
	printMeasurement("zap 强类型字段", time.Since(startedAt))
}

// measureSlog 测量 slog 属性字段的调用耗时。
func measureSlog(request Request) {
	logger := NewSlogLogger(measurementOutput)
	logger.LogAttrs(nil, slog.LevelInfo, "request completed",
		slog.String("request_id", request.ID),
		slog.String("path", request.Path),
		slog.Int("status", request.Status),
		slog.Duration("latency", request.Latency),
	)

	startedAt := time.Now()
	for range measurementIterations {
		logger.LogAttrs(nil, slog.LevelInfo, "request completed",
			slog.String("request_id", request.ID),
			slog.String("path", request.Path),
			slog.Int("status", request.Status),
			slog.Duration("latency", request.Latency),
		)
	}
	printMeasurement("slog 属性字段", time.Since(startedAt))
}

// measureStdLog 测量标准库 log.Printf 的调用耗时。
func measureStdLog(request Request) {
	logger := NewStdLogger(measurementOutput)
	logger.Printf("request completed request_id=%s path=%s status=%d latency=%s",
		request.ID, request.Path, request.Status, request.Latency)

	startedAt := time.Now()
	for range measurementIterations {
		logger.Printf("request completed request_id=%s path=%s status=%d latency=%s",
			request.ID, request.Path, request.Status, request.Latency)
	}
	printMeasurement("log.Printf", time.Since(startedAt))
}

// printMeasurement 输出总耗时和单次平均耗时。
func printMeasurement(name string, elapsed time.Duration) {
	average := float64(elapsed) / float64(measurementIterations)
	fmt.Printf("%s：总耗时 %s，平均 %.1f ns/op\n", name, elapsed, average)
}
