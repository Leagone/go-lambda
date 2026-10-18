package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"lambda/internal/domain"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MessageProcessor interface {
	Process(ctx context.Context, event domain.Event) error
}

type HttpProcessor struct {
	client  HttpClient
	log     *slog.Logger
	baseUrl *url.URL
}

func (p *HttpProcessor) Process(ctx context.Context, event domain.Event) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	traceID := spanCtx.TraceID().String()
	spanID := spanCtx.SpanID().String()
	ipStr := event.IPAddress.String()

	reqUrl := p.baseUrl.JoinPath(ipStr)
	reqUrl.RawQuery = url.Values{
		"fields": {"city,continent,location"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	p.log.InfoContext(ctx, "fetching location data",
		slog.String("trace_id", traceID),
		slog.String("span_id", spanID),
		slog.String("ip", ipStr),
	)

	var res *http.Response
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		res, err = p.client.Do(req)
		if err == nil && res.StatusCode < 500 {
			break
		}
		if res != nil {
			res.Body.Close()
		}

		p.log.WarnContext(ctx, "transient error fetching location, retrying",
			slog.String("trace_id", traceID),
			slog.String("span_id", spanID),
			slog.String("ip", ipStr),
			slog.Int("attempt", attempt+1),
			slog.Any("error", err),
		)

		backoff := time.Duration(1<<attempt) * time.Second

		select {
		case <-ctx.Done():
			return fmt.Errorf("context canceled during retry: %w", ctx.Err())
		case <-time.After(backoff):
		}
	}

	if err != nil {
		return fmt.Errorf("executing request for IP %s: %w", ipStr, err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("API error for IP %s: status %d", ipStr, res.StatusCode)
	}

	p.log.InfoContext(ctx, "retrieved location data",
		slog.String("trace_id", traceID),
		slog.String("span_id", spanID),
		slog.String("ip", ipStr),
		slog.String("status", res.Status),
	)

	var location domain.Location
	if err := json.NewDecoder(res.Body).Decode(&location); err != nil {
		return fmt.Errorf("decoding API response for IP %s: %w", ipStr, err)
	}

	p.log.InfoContext(ctx, "successfully decoded location data",
		slog.String("trace_id", traceID),
		slog.String("span_id", spanID),
		slog.String("ip", ipStr),
		slog.Any("location", location),
	)

	return nil
}

func NewProcessor(client HttpClient, log *slog.Logger, baseUrl *url.URL) *HttpProcessor {
	return &HttpProcessor{
		client:  client,
		log:     log,
		baseUrl: baseUrl,
	}
}
