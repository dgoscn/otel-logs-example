# otel-logs-example: Structured Logs with OpenTelemetry SDK in Go

**Title:** Generating and Exporting Structured Logs with the OpenTelemetry SDK

**Author:** Diego Amaral
**Date:** 2025-06-05

---

## Overview

This repository demonstrates how to generate, process, and export structured logs using the OpenTelemetry SDK in Go. It focuses on log instrumentation, SDK setup, and local debugging, while also enabling log enrichment through the OpenTelemetry Collector.

---

## 1 - Objective

Instrument a Go application using the OpenTelemetry SDK to emit structured logs and metrics, simulate synthetic log events, and export enriched telemetry to a backend such as New Relic.

### 1.1 Scope

* **Instrumentation:** Emit structured logs with `otelzap`, and record metrics using `otel.Meter`.
* **Data Enrichment:** Enrich logs via the OpenTelemetry Collector using a GeoIP processor.
* **Export:** Send enriched logs to New Relic via `otlphttp` or debug them locally.
* **Tools:** This repo complements a custom OpenTelemetry Collector created with OCDB.

---

## 2 - Project Structure

```
otel-logs-example/
├── cmd/web-based/       # Main application entrypoint (main.go)
├── internal/telemetry/  # SDK setup logic (otel.go)
├── otel.yaml            # Configuration file for the OpenTelemetry SDK
├── go.mod, go.sum       # Module dependencies
```

---

## 3 - Application Instrumentation

### 3.1 SDK Initialization

Defined in `internal/telemetry/otel.go`, the `Setup` function reads `otel.yaml`, initializes the SDK, and configures the global tracer, meter, and logger providers:

```go
func Setup(ctx context.Context, confFlag string) (func(context.Context) error, error) {
  b, err := os.ReadFile(confFlag)
  // Parse config and start SDK...
  otel.SetTracerProvider(sdk.TracerProvider())
  otel.SetMeterProvider(sdk.MeterProvider())
  global.SetLoggerProvider(sdk.LoggerProvider())
  return sdk.Shutdown, nil
}
```

---

### 3.2 Main Application Logic

The application uses a ticker to emit log events every minute. Each event contains simulated user data:

```go
meter := otel.Meter("otel-logs-example")
requestsCounter, _ := meter.Int64Counter("requests")
requestDurationHist, _ := meter.Float64Histogram("log_duration_seconds")

ticker := time.NewTicker(1 * time.Minute)
for range ticker.C {
    start := time.Now()
    requestsCounter.Add(ctx, 1)
    requestDurationHist.Record(ctx, time.Since(start).Seconds())
    logger.Info("New session log event", zap.String("data", string(eventJson)))
}
```

Logs are emitted in JSON format with a `data` field containing a JSON string of structured information.

---

## 4 - OpenTelemetry SDK Configuration (otel.yaml)

```yaml
file_format: "0.3"
disabled: false
meter_provider:
  readers:
    - periodic:
        interval: 1000
        exporter:
          otlp:
            protocol: http/protobuf
            endpoint: http://localhost:4318

logger_provider:
  processors:
    - batch:
        exporter:
          otlp:
            protocol: http/protobuf
            endpoint: http://localhost:4318
```

This configuration enables the SDK to batch and export logs and metrics to the local OpenTelemetry Collector via OTLP/HTTP.

---

## 5 - Running the Application

```sh
cd otel-logs-example

go mod tidy
go run ./cmd/web-based
```

Expected output:

```json
{"level":"info","ts":1739921689.809112,"msg":"starting the ticker server"}
{"level":"info","ts":1739921690.809581,"msg":"New session log event","data":"{...}"}
```

If you see connection errors like `connection refused`, make sure the OpenTelemetry Collector is running and listening on port `4318`.

---

## 6 - Observability Backends

This project supports multiple debugging or export options:

* **otel-tui**: Great for local observability visualization.
* **New Relic**: Used as the final log sink for real-world scenarios.
* **debug exporter**: Visualizes logs directly in the terminal for validation.

All exporters rely on the Collector configuration. See the main blog post or repository for `geoip-collector` for more details.

---

## Final Notes

This example aims to bridge synthetic log generation with end-to-end OpenTelemetry log pipelines. It demonstrates how to emit, structure, and enrich logs in real-time for production-ready observability flows.

If you're running this with the custom collector (`geoip-collector`), make sure both are operating concurrently for complete functionality.

Happy hacking!
