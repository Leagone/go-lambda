# Go Lambda

A small AWS Lambda written in Go as a learning project to explore the Go language, AWS Lambda integration, OpenTelemetry, and event-driven processing.

This project listens for SQS messages, reads an `ip_address` from each message body, calls a geolocation API for that IP, and logs the result. It is intentionally simple and educational rather than a fully production-hardened service.

## What this project does

The Lambda receives an SQS event and processes each record:

1. Reads a message body like:
   ```json
   {"ip_address": "192.168.1.1"}
   ```
2. Parses the payload into a Go struct.
3. Calls a configured geolocation API using the IP address.
4. Requests a subset of location fields (`city`, `continent`, and `location`).
5. Logs the response through Go's structured logger (`slog`).
6. Emits tracing metadata using OpenTelemetry.

## Architecture

```text
cmd/
  sqs-processor/
    main.go                  Application entry point

internal/
  domain/
    domain.go                Core event and API response models
  handler/
    handler.go               SQS event handler
    sqs_carrier.go           Trace propagation helper for SQS messages
    handler_test.go          Unit tests for the handler
    integration_test.go      Integration test with an HTTP mock server
  processor/
    processor.go             Business logic for the API request
    processors_test.go       Processor tests
  telemetry/
    tracer.go                OpenTelemetry initialization logic
```

### Main flow

- `cmd/sqs-processor/main.go` sets up the Lambda runtime, tracing, HTTP client, and processor.
- `internal/handler/handler.go` extracts the SQS message body, unmarshals it, and calls the processor.
- `internal/processor/processor.go` executes an HTTP GET request to the configured geolocation endpoint and decodes the response.
- `internal/domain/domain.go` defines the input event and geolocation output shape.

## Prerequisites

- Go 1.27+
- AWS Lambda runtime compatible with Go
- Access to an HTTP API that can resolve IP geolocation data

## Configuration

The application expects the following environment variable:

- `API_BASE_URL` — base URL for the external IP location API

Example:

```bash
export API_BASE_URL="https://api.example.com"
```

This project builds the request as:

```text
{API_BASE_URL}/{ip}?fields=city,continent,location
```

## Local development

### Install dependencies

```bash
go mod download
```

### Run tests

```bash
go test ./...
```

### Build the binary

```bash
go build ./...
```

For AWS Lambda deployment, the binary typically needs to be built for Linux and named `bootstrap`:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bootstrap ./cmd/sqs-processor
```

Then package it as a zip file for deployment to Lambda.

## Example SQS Event

```json
{
  "Records": [
    {
      "messageId": "abc123",
      "body": "{\"ip_address\":\"192.168.1.1\"}",
      "messageAttributes": {
        "traceparent": {
          "stringValue": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01",
          "dataType": "String"
        }
      }
    }
  ]
}
```

## Dependencies

This project uses a few Go libraries to support learning and experimentation:

- `github.com/aws/aws-lambda-go` — AWS Lambda Go runtime
- `go.opentelemetry.io/otel` — tracing and telemetry
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` — HTTP instrumentation
- `log/slog` — structured logging in the standard library

## Learning notes

This repository is intentionally focused on learning and understanding a practical Go Lambda workflow:

- building a Lambda handler for SQS events
- parsing JSON payloads
- calling external REST APIs
- applying structured logging
- tracing request flow with OpenTelemetry
- organizing code using `cmd/` and `internal/`

## Limitations

This is not a production-ready, full enterprise service. It is best understood as a learning project and a proof-of-concept for:

- Go + AWS Lambda integration
- SQS-based event processing
- external API lookup patterns
- request tracing and observability basics

## Ideas for extension

If you want to keep learning, good next steps would be:

- add retries and retry policies for transient failures
- add error handling and dead-letter queues
- validate configuration and environment variables
- add CI with `go test`, `go vet`, and linting
- add deployment via Terraform, SAM, or CDK
- add metrics and more complete tracing around downstream service calls

## License

This project is provided as a learning exercise. Check the repository for any licensing information if you plan to reuse or extend it.

---

This project is meant to be simple, readable, and educational. It showcases how a Go Lambda can listen to SQS, transform event data, and make outbound API calls in a clean package layout.
