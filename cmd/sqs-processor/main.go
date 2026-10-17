package main

import (
	"context"
	"lambda/internal/config"
	"lambda/internal/handler"
	"lambda/internal/processor"
	"lambda/internal/telemetry"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

func main() {

	ctx := context.Background()
	shutdown, err := telemetry.InitTracer(ctx, "test-lambda")

	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}

	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Printf("failed to shutdown tracer safely: %v", err)
		}
	}()

	cfg, err := config.NewConfig()

	if err != nil {
		log.Fatalf("failed to load lambda config %v", err)
	}

	tracer := otel.Tracer("sqs-lambda")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	msgProcessor := processor.NewProcessor(httpClient, logger, &cfg.ApiUrl)
	sqsHandler := handler.NewHandler(msgProcessor, tracer)

	lambda.Start(sqsHandler.HandleSQSEvent)
}
