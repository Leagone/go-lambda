package handler_test

import (
	"context"
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
		name        string
		mockError   error
		expectError bool
	}{
		{
			name:        "Happy Path",
			mockError:   nil,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tracer := otel.Tracer("test")
			processor := &MockProcessor{ReturnError: tc.mockError}
			handler := handler.NewHandler(processor, tracer, 5*time.Second)
			ctx := context.Background()
			event := events.SQSEvent{
				Records: []events.SQSMessage{
					{
						Body: `{"ip_address": "192.168.1.1"}`,
					},
				},
			}
			err := handler.HandleSQSEvent(ctx, event)

			if tc.expectError && err == nil {
				t.Errorf("expected an error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}
		})
	}
}
