package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const measurementIterations = 1_000_000

// Request 是三个日志系统共用的结构化数据。
type Request struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Status int    `json:"status"`
}

// primitiveRecord 是标准库 log 生成等价 JSON 时使用的记录结构。
type primitiveRecord struct {
	Message   string `json:"msg"`
	RequestID string `json:"request_id"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
}

// sampleRequest 返回三种日志系统共用的固定输入。
func sampleRequest() Request {
	return Request{
		ID:     "req-9f5d20",
		Path:   "/v1/orders/42",
		Status: 200,
	}
}

// NewZapLogger 创建不包含时间和级别字段的 JSON 日志器。
func NewZapLogger(out io.Writer) *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = ""
	config.LevelKey = ""
	encoder := zapcore.NewJSONEncoder(config)
	core := zapcore.NewCore(encoder, zapcore.AddSync(out), zap.InfoLevel)
	return zap.New(core)
}

// NewSlogLogger 创建不包含时间和级别字段的 JSON 日志器。
func NewSlogLogger(out io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 && (attr.Key == slog.TimeKey || attr.Key == slog.LevelKey) {
				return slog.Attr{}
			}
			return attr
		},
	})
	return slog.New(handler)
}

// NewStdLogger 创建不带前缀和时间标记的标准库日志器。
func NewStdLogger(out io.Writer) *log.Logger {
	return log.New(out, "", 0)
}

// writeStdJSON 让标准库 log 完成与 zap、slog 等价的 JSON 输出。
func writeStdJSON(logger *log.Logger, record any) {
	encoded, err := json.Marshal(record)
	if err != nil {
		panic(fmt.Sprintf("编码标准库日志失败：%v", err))
	}
	logger.Print(string(encoded))
}

// discardWriter 最终写入 io.Discard，但不会触发标准库 log 的快速跳过逻辑。
type discardWriter struct {
	target io.Writer
}

func (w discardWriter) Write(p []byte) (int, error) {
	return w.target.Write(p)
}

var discardedOutput io.Writer = discardWriter{target: io.Discard}

// main 先展示等价 JSON，再横向测量三种日志系统的运行耗时。
func main() {
	request := sampleRequest()
	showLogExamples(request)

	fmt.Printf("\n等价 JSON 日志耗时：每种日志系统记录 %d 次\n", measurementIterations)
	measureZap(request)
	measureSlog(request)
	measureStdLog(request)
}

// showLogExamples 输出三条语义相同的 JSON 日志。
func showLogExamples(request Request) {
	standard := NewStdLogger(os.Stdout)
	writeStdJSON(standard, primitiveRecord{
		Message:   "request completed",
		RequestID: request.ID,
		Path:      request.Path,
		Status:    request.Status,
	})

	structured := NewSlogLogger(os.Stdout)
	structured.LogAttrs(nil, slog.LevelInfo, "request completed",
		slog.String("request_id", request.ID),
		slog.String("path", request.Path),
		slog.Int("status", request.Status),
	)

	fast := NewZapLogger(os.Stdout)
	fast.Info("request completed",
		zap.String("request_id", request.ID),
		zap.String("path", request.Path),
		zap.Int("status", request.Status),
	)
	_ = fast.Sync()
}

// measureZap 测量 zap 生成等价 JSON 的耗时。
func measureZap(request Request) {
	logger := NewZapLogger(discardedOutput)
	logger.Info("request completed",
		zap.String("request_id", request.ID),
		zap.String("path", request.Path),
		zap.Int("status", request.Status),
	)

	startedAt := time.Now()
	for range measurementIterations {
		logger.Info("request completed",
			zap.String("request_id", request.ID),
			zap.String("path", request.Path),
			zap.Int("status", request.Status),
		)
	}
	printMeasurement("zap", time.Since(startedAt))
}

// measureSlog 测量 slog 生成等价 JSON 的耗时。
func measureSlog(request Request) {
	logger := NewSlogLogger(discardedOutput)
	logger.LogAttrs(nil, slog.LevelInfo, "request completed",
		slog.String("request_id", request.ID),
		slog.String("path", request.Path),
		slog.Int("status", request.Status),
	)

	startedAt := time.Now()
	for range measurementIterations {
		logger.LogAttrs(nil, slog.LevelInfo, "request completed",
			slog.String("request_id", request.ID),
			slog.String("path", request.Path),
			slog.Int("status", request.Status),
		)
	}
	printMeasurement("slog", time.Since(startedAt))
}

// measureStdLog 测量标准库 log 生成等价 JSON 的耗时。
func measureStdLog(request Request) {
	logger := NewStdLogger(discardedOutput)
	record := primitiveRecord{
		Message:   "request completed",
		RequestID: request.ID,
		Path:      request.Path,
		Status:    request.Status,
	}
	writeStdJSON(logger, record)

	startedAt := time.Now()
	for range measurementIterations {
		writeStdJSON(logger, record)
	}
	printMeasurement("log + encoding/json", time.Since(startedAt))
}

// printMeasurement 输出总耗时和单次平均耗时。
func printMeasurement(name string, elapsed time.Duration) {
	average := float64(elapsed) / float64(measurementIterations)
	fmt.Printf("%s：总耗时 %s，平均 %.1f ns/op\n", name, elapsed, average)
}
