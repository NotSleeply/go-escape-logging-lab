package main

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"go.uber.org/zap"
)

var benchmarkError = errors.New("database unavailable")

// opaqueDiscard 转发到 io.Discard，但自身不等于 io.Discard。
// 标准库 log 识别到直接传入 io.Discard 时会在格式化参数前直接返回，
// 会使三者的对比失去意义。
type opaqueDiscard struct {
	target io.Writer
}

func (w opaqueDiscard) Write(p []byte) (int, error) {
	return w.target.Write(p)
}

var benchmarkOutput io.Writer = opaqueDiscard{target: io.Discard}

// 三个日志器均启用，最终写入 io.Discard。这样会保留编码和参数处理路径，
// 同时排除终端和磁盘 I/O 对结果的影响。
func BenchmarkPrimitiveFields(b *testing.B) {
	request := SampleRequest()

	b.Run("zap_typed", func(b *testing.B) {
		logger := NewZapLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Info("request completed",
				zap.String("request_id", request.ID),
				zap.String("path", request.Path),
				zap.Int("status", request.Status),
			)
		}
	})

	b.Run("slog_attrs", func(b *testing.B) {
		logger := NewSlogLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.LogAttrs(nil, slog.LevelInfo, "request completed",
				slog.String("request_id", request.ID),
				slog.String("path", request.Path),
				slog.Int("status", request.Status),
			)
		}
	})

	b.Run("log_printf", func(b *testing.B) {
		logger := NewStdLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Printf("request completed request_id=%s path=%s status=%d",
				request.ID, request.Path, request.Status)
		}
	})
}

func BenchmarkStructuredValue(b *testing.B) {
	b.Run("zap_any", func(b *testing.B) {
		logger := NewZapLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			request := SampleRequest()
			logger.Info("request completed", zap.Any("request", request))
		}
	})

	b.Run("slog_any", func(b *testing.B) {
		logger := NewSlogLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			request := SampleRequest()
			logger.Info("request completed", slog.Any("request", request))
		}
	})

	b.Run("log_printf", func(b *testing.B) {
		logger := NewStdLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			request := SampleRequest()
			logger.Printf("request completed request=%+v", request)
		}
	})
}

func BenchmarkStructuredPointer(b *testing.B) {
	b.Run("zap_any", func(b *testing.B) {
		logger := NewZapLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			request := SampleRequest()
			logger.Info("request completed", zap.Any("request", &request))
		}
	})

	b.Run("slog_any", func(b *testing.B) {
		logger := NewSlogLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			request := SampleRequest()
			logger.Info("request completed", slog.Any("request", &request))
		}
	})

	b.Run("log_printf", func(b *testing.B) {
		logger := NewStdLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			request := SampleRequest()
			logger.Printf("request completed request=%+v", &request)
		}
	})
}

func BenchmarkErrorField(b *testing.B) {
	b.Run("zap_error", func(b *testing.B) {
		logger := NewZapLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Info("request failed", zap.Error(benchmarkError))
		}
	})

	b.Run("slog_any", func(b *testing.B) {
		logger := NewSlogLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Info("request failed", slog.Any("error", benchmarkError))
		}
	})

	b.Run("log_printf", func(b *testing.B) {
		logger := NewStdLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Printf("request failed error=%v", benchmarkError)
		}
	})
}

func BenchmarkDynamicArguments(b *testing.B) {
	request := SampleRequest()

	b.Run("zap_sugared", func(b *testing.B) {
		logger := NewZapLogger(benchmarkOutput).Sugar()
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Infow("request completed",
				"request_id", request.ID,
				"path", request.Path,
				"status", request.Status,
			)
		}
	})

	b.Run("slog_key_values", func(b *testing.B) {
		logger := NewSlogLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Info("request completed",
				"request_id", request.ID,
				"path", request.Path,
				"status", request.Status,
			)
		}
	})

	b.Run("log_printf", func(b *testing.B) {
		logger := NewStdLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Printf("request completed request_id=%s path=%s status=%d",
				request.ID, request.Path, request.Status)
		}
	})
}
