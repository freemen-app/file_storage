package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func ConfigureLogger(level string) {
	levels := map[string]zerolog.Level{
		"debug":   zerolog.DebugLevel,
		"info":    zerolog.InfoLevel,
		"warning": zerolog.WarnLevel,
		"error":   zerolog.ErrorLevel,
	}
	level = strings.ToLower(level)

	logLevel, ok := levels[level]
	if !ok {
		panic(fmt.Sprintf("unknown level: %s", level))
	}

	zerolog.New(os.Stdout)
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(logLevel)
}
