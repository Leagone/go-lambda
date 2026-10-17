package handler_test

import (
	"context"
	"lambda/internal/handler"
	"lambda/internal/processor"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func TestIntegration_Pipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	mockAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fields") != "city,continent,location" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ip": "192.168.1.1", "country": "US", "city": "Integration City"}`))
	}))
	defer mockAPIServer.Close()

	apiUrl, _ := url.Parse(mockAPIServer.URL)

	tracer := otel.Tracer("test")
	slog := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	processor := processor.NewProcessor(httpClient, slog, apiUrl)
	handler := handler.NewHandler(processor, tracer)

	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "mock-message-id",
				Body:      `{"ip_address": "192.168.1.1"}`,
				MessageAttributes: map[string]events.SQSMessageAttribute{
					"traceparent": {
						DataType:    "String",
						StringValue: aws.String("00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"),
					},
				},
			},
		},
	}

	err := handler.HandleSQSEvent(ctx, event)

	if err != nil {
		t.Errorf("handler pipeline failed: %v", err)
	}

}
