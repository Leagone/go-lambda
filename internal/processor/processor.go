package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"lambda/internal/domain"
	"log/slog"
	"net/http"
	"net/url"
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

	reqUrl := p.baseUrl.JoinPath(event.IPAddress.String())

	reqUrl.RawQuery = url.Values{
		"fields": {"city,continent,location"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)

	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	p.log.InfoContext(ctx, "fetching location data for", "ip", event.IPAddress.String())

	res, err := p.client.Do(req)

	if err != nil {
		return fmt.Errorf("executing request for IP %s: %w", event.IPAddress.String(), err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("API error for IP %s: status %d", event.IPAddress.String(), res.StatusCode)
	}

	p.log.InfoContext(ctx, "retreived location data for", "ip", res.Status)

	var location domain.Location

	if err := json.NewDecoder(res.Body).Decode(&location); err != nil {
		return fmt.Errorf("decoding API response for IP %s: %w", event.IPAddress.String(), err)
	}

	p.log.InfoContext(ctx, "successfully decoded location data",
		"ip", event.IPAddress.String(),
		"location", location,
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
