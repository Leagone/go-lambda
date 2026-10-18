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
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
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
			name: "Processor successfully returns 200 and Location for given IP",
			mockClient: &MockHttpClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString(`{"ip":"81.110.173.213","country":"GB","city":"Woolwich","continent":"EU","location":{"latitude":51.4966,"longitude":0.0666,"accuracy_radius":10,"time_zone":"Europe/London"}}`)),
					}, nil
				},
			},
			expectErr: false,
		},
		{
			name: "Processor fails and returns 400",
			mockClient: &MockHttpClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(bytes.NewBufferString(`{"error": "ip not found"}`)),
					}, nil
				},
			},
			expectErr: true,
		},
		{
			name: "Processor retries on 500 and succeeds on attempt 3",
			mockClient: &MockHttpClient{
				DoFunc: func() func(req *http.Request) (*http.Response, error) {
					attempts := 0
					return func(req *http.Request) (*http.Response, error) {
						attempts++
						if attempts < 3 {
							return &http.Response{
								StatusCode: 500,
								Body:       io.NopCloser(bytes.NewBufferString(`Internal Server Error`)),
							}, nil
						}
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewBufferString(`{"ip":"81.110.173.213","country":"GB","city":"Woolwich","continent":"EU","location":{"latitude":51.4966,"longitude":0.0666,"accuracy_radius":10,"time_zone":"Europe/London"}}`)),
						}, nil
					}
				}(),
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			p := processor.NewProcessor(tc.mockClient, logger, testURL)
			event := domain.Event{
				IPAddress: []byte{81, 110, 173, 213},
			}

			err := p.Process(ctx, event)

			if tc.expectErr && err == nil {
				t.Errorf("expected an error but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("did not expect error but got: %v", err)
			}

		})
	}
}
