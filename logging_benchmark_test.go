package main

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"go.uber.org/zap"
)

var benchmarkError = errors.New("database unavailable")

// opaqueDiscard forwards to io.Discard without being equal to io.Discard.
// The standard log package recognizes a direct io.Discard writer and returns
// before formatting its arguments, which would make this comparison invalid.
type opaqueDiscard struct {
	target io.Writer
}

func (w opaqueDiscard) Write(p []byte) (int, error) {
	return w.target.Write(p)
}

var benchmarkOutput io.Writer = opaqueDiscard{target: io.Discard}

// Each logger is enabled and ultimately writes to io.Discard. This deliberately
// keeps encoding and argument handling on the measured path while removing
// terminal and disk I/O from the result.
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
				zap.Duration("latency", request.Latency),
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
				slog.Duration("latency", request.Latency),
			)
		}
	})

	b.Run("log_printf", func(b *testing.B) {
		logger := NewStdLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Printf("request completed request_id=%s path=%s status=%d latency=%s",
				request.ID, request.Path, request.Status, request.Latency)
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
				"latency", request.Latency,
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
				"latency", request.Latency,
			)
		}
	})

	b.Run("log_printf", func(b *testing.B) {
		logger := NewStdLogger(benchmarkOutput)
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			logger.Printf("request completed request_id=%s path=%s status=%d latency=%s",
				request.ID, request.Path, request.Status, request.Latency)
		}
	})
}
