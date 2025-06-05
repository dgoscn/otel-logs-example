package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"otel-logs-example/internal/telemetry"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SessionEvent represents a synthetic user session activity
type SessionEvent struct {
	SessionID     string `json:"session_id"`
	CustomerEmail string `json:"email"`
	LoginCountry  string `json:"country"`
	Browser       string `json:"browser"`
	LoginTime     string `json:"login_time"`
	IPAddress     string `json:"ip_address"`
}

func generateSessionEvent() SessionEvent {
	return SessionEvent{
		SessionID:     gofakeit.UUID(),
		CustomerEmail: gofakeit.Email(),
		LoginCountry:  gofakeit.Country(),
		Browser:       gofakeit.UserAgent(),
		LoginTime:     time.Now().Format(time.RFC3339),
		IPAddress:     gofakeit.IPv4Address(),
	}
}

func main() {
	otelConfigFlag := flag.String("otel", "./otel.yaml", "Path to OpenTelemetry config")
	flag.Parse()

	// Structured + OTel logger
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		otelzap.NewCore("otel-logs-example", otelzap.WithLoggerProvider(global.GetLoggerProvider())),
	)
	logger := zap.New(core)

	// Initialize OTel SDK
	closer, err := telemetry.Setup(context.Background(), *otelConfigFlag)
	if err != nil {
		logger.Fatal("Failed to setup telemetry SDK", zap.Error(err))
	}
	defer closer(context.Background())

	// Meters
	meter := otel.Meter("otel-logs-example")
	requestsCounter, _ := meter.Int64Counter("session_requests_total")
	errorCounter, _ := meter.Int64Counter("session_errors_total")
	durationHist, _ := meter.Float64Histogram("session_processing_duration_seconds")

	logger.Info("Session generator started")

	ticker := time.NewTicker(15 * time.Second)
	for range ticker.C {
		start := time.Now()
		ctx := context.Background()

		// Generate event
		event := generateSessionEvent()
		payload, err := json.Marshal(event)
		if err != nil {
			logger.Error("Error serializing session event", zap.Error(err))
			errorCounter.Add(ctx, 1)
			continue
		}

		// Log structured event via OTel logger
		logger.Info("New session log event", zap.String("data", string(payload)))

		// Metrics
		requestsCounter.Add(ctx, 1)
		durationHist.Record(ctx, time.Since(start).Seconds())
	}
}
