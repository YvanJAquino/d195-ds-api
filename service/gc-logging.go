package main

import (
	"context"

	"cloud.google.com/go/logging"
)

type GoogleCloudLogger struct {
	*logging.Client
	logger *logging.Logger
}

func NewGoogleCloudLogger(ctx context.Context, parent string) (*GoogleCloudLogger, error) {
	client, err := logging.NewClient(ctx, parent)
	if err != nil {
		return nil, err
	}
	logger := new(GoogleCloudLogger)
	logger.Client = client
	return logger, nil
}

func (l *GoogleCloudLogger) SetLogger(logID string, opts ...logging.LoggerOption) func() error {
	l.logger = l.Logger(logID)
	return l.logger.Flush
}

func (l *GoogleCloudLogger) Log(payload any, severity logging.Severity) {
	entry := logging.Entry{
		Payload:  payload,
		Severity: severity,
	}
	l.logger.Log(entry)
}

// Convenience type Aliases

// const name = "log-example"
// logger := client.Logger(name)
// defer logger.Flush() // Ensure the entry is written.

// logger.Log(logging.Entry{
//         // Log anything that can be marshaled to JSON.
//         Payload: struct{ Anything string }{
//                 Anything: "The payload can be any type!",
//         },
//         Severity: logging.Debug,
// })
