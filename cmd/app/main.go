package main

import (
	"fmt"
	"log"

	"github.com/realPointer/banners/pkg/logger"
)

func main() {
	l, err := logger.New("debug")
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create logger: %w", err))
	}

	l.Info().Msg("Hello, World!")
}
