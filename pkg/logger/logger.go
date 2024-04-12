package logger

import (
	"fmt"

	"github.com/rs/zerolog"
)

func New(level string) (*zerolog.Logger, error) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	consoleWriter := zerolog.NewConsoleWriter()
	logger := zerolog.New(consoleWriter).Level(l).
		With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount).
		Logger()

	return &logger, nil
}
