package handler_test

import (
	"context"
	"errors"
	"lambda/internal/domain"
	"lambda/internal/handler"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"go.opentelemetry.io/otel"
)

type MockProcessor struct {
	ReturnError error
}

func (moc *MockProcessor) Process(ctx context.Context, event domain.Event) error {
	return moc.ReturnError
}

func TestHandleSQSEvent(t *testing.T) {
	tests := []struct {
		name             string
		mockError        error
		expectedFailures int
	}{
		{
			name:             "Happy Path - zero failures",
			mockError:        nil,
			expectedFailures: 0,
		},
		{
			name:             "Processor Error - returns partial batch failure",
			mockError:        errors.New("mock processing error"),
			expectedFailures: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tracer := otel.Tracer("test")
			meter := otel.Meter("test")

			processor := &MockProcessor{ReturnError: tc.mockError}
			h := handler.NewHandler(processor, tracer, meter, 5*time.Second)

			ctx := context.Background()
			event := events.SQSEvent{
				Records: []events.SQSMessage{
					{
						MessageId: "test-msg-123",
						Body:      `{"ip_address": "192.168.1.1"}`,
					},
				},
			}

			resp, err := h.HandleSQSEvent(ctx, event)

			if err != nil {
				t.Fatalf("expected handler error to always be nil, got: %v", err)
			}

			if len(resp.BatchItemFailures) != tc.expectedFailures {
				t.Errorf("expected %d batch item failures, got %d", tc.expectedFailures, len(resp.BatchItemFailures))
			}

			if tc.expectedFailures > 0 && resp.BatchItemFailures[0].ItemIdentifier != "test-msg-123" {
				t.Errorf("expected failed message ID 'test-msg-123', got '%s'", resp.BatchItemFailures[0].ItemIdentifier)
			}
		})
	}
}
