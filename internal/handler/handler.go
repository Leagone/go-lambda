package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"lambda/internal/domain"
	"lambda/internal/processor"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	messageProcessor processor.MessageProcessor
	tracer           trace.Tracer
	timeOut time.Duration
}

func (h *Handler) HandleSQSEvent(ctx context.Context, event events.SQSEvent) error {

	for _, record := range event.Records {
		err := h.processRecord(ctx, record)
		if err != nil {
			return fmt.Errorf("processing record failed: %w", err)
		}
	}

	return nil
}

func (h *Handler) processRecord(ctx context.Context, record events.SQSMessage) error {

	ctx, cancel := context.WithTimeout(ctx, h.timeOut)
    defer cancel()

	carrier := SQSCarrier(record.MessageAttributes)
	propagator := propagation.TraceContext{}

	ctx = propagator.Extract(ctx, carrier)

	ctx, span := h.tracer.Start(ctx, "process_record")
	defer span.End()

	var event domain.Event
	if err := json.Unmarshal([]byte(record.Body), &event); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to unmarshall the body")
		return fmt.Errorf("unmarshaling record body: %w", err)
	}

	if err := h.messageProcessor.Process(ctx, event); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to process the record")
		return fmt.Errorf("processing event: %w", err)
	}

	return nil
}

func NewHandler(processor processor.MessageProcessor, tracer trace.Tracer, requestTimeout time.Duration) *Handler {
	return &Handler{
		messageProcessor: processor,
		tracer:           tracer,
		timeOut: requestTimeout,
	}
}
