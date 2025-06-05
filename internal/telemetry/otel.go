package telemetry

import (
	"context"
	"os"

	config "go.opentelemetry.io/contrib/config/v0.3.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
)

// Meter is a global meter instance
var Meter = otel.GetMeterProvider().Meter("geoip-meter")

// Setup initializes the OpenTelemetry SDK with the given configuration
func Setup(ctx context.Context, confFlag string) (func(context.Context) error, error) {
	// read the configuration file
	b, err := os.ReadFile(confFlag)
	if err != nil {
		return nil, err
	}

	// interpolate the environment variables
	b = []byte(os.ExpandEnv(string(b)))

	// parse the config
	conf, err := config.ParseYAML(b)
	if err != nil {
		return nil, err
	}

	// Create a new SDK instance with the parsed configuration
	sdk, err := config.NewSDK(config.WithContext(ctx), config.WithOpenTelemetryConfiguration(*conf))
	if err != nil {
		return nil, err
	}

	// Set the global text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// Set the global tracer provider
	otel.SetTracerProvider(sdk.TracerProvider())

	// Set the global meter provider
	otel.SetMeterProvider(sdk.MeterProvider())

	// Set the global logger provider
	global.SetLoggerProvider(sdk.LoggerProvider())

	// Return the SDK shutdown function
	return sdk.Shutdown, nil
}
