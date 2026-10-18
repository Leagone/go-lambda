package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"lambda/internal/domain"
	"lambda/internal/processor"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	messageProcessor processor.MessageProcessor
	tracer           trace.Tracer
	timeOut          time.Duration
	msgCounter       metric.Int64Counter
}

func (h *Handler) HandleSQSEvent(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	var response events.SQSEventResponse

	for _, record := range event.Records {
		err := h.processRecord(ctx, record)
		if err != nil {
			response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
		}
	}

	return response, nil
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
		h.msgCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("status", "error")))
		return fmt.Errorf("unmarshaling record body: %w", err)
	}

	err := h.messageProcessor.Process(ctx, event)

	status := "success"
	if err != nil {
		status = "error"
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to process the record")
	}

	h.msgCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("status", status)))

	if err != nil {
		return fmt.Errorf("processing event: %w", err)
	}

	return nil
}

func NewHandler(processor processor.MessageProcessor, tracer trace.Tracer, meter metric.Meter, requestTimeout time.Duration) *Handler {
	counter, _ := meter.Int64Counter("sqs_messages_processed_total")
	return &Handler{
		messageProcessor: processor,
		tracer:           tracer,
		timeOut:          requestTimeout,
		msgCounter:       counter,
	}
}
