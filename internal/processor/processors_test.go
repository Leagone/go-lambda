package processor_test

import (
	"bytes"
	"context"
	"io"
	"lambda/internal/domain"
	"lambda/internal/processor"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
)

type MockHttpClient struct {
	ResponseStatusCode int
	ResponseBody       string
	Err                error
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	return &http.Response{
		StatusCode: m.ResponseStatusCode,
		Body:       io.NopCloser(bytes.NewBufferString(m.ResponseBody)),
	}, nil
}

func TestHttpProcessor_process(t *testing.T) {

	ctx := context.Background()
	logger := slog.New(slog.DiscardHandler)
	testURL, _ := url.Parse("https://api.example.com")

	tests := []struct {
		name       string
		mockClient *MockHttpClient
		expectErr  bool
	}{
		{
			name: "Processor succesfully return 200 and Location for given IP",
			mockClient: &MockHttpClient{
				ResponseStatusCode: 200,
				ResponseBody:       `{"ip":"81.110.173.213","country":"GB","city":"Woolwich","continent":"EU","location":{"latitude":51.4966,"longitude":0.0666,"accuracy_radius":10,"time_zone":"Europe/London"}}`,
			},
			expectErr: false,
		},
		{
			name: "Processor fails and returns 404",
			mockClient: &MockHttpClient{
				ResponseStatusCode: 400,
				ResponseBody:       `{"error": "ip not found"}`,
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			processor := processor.NewProcessor(tc.mockClient, logger, testURL)
			event := domain.Event{
				IPAddress: []byte{8, 8, 8, 8},
			}

			err := processor.Process(ctx, event)

			if tc.expectErr && err == nil {
				t.Errorf("expected an error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("did not expect error but got :%v", err)
			}

		})
	}

}
