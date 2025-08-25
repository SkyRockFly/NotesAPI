package applogger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoggerCfg struct {
	FormatTimestamp string
	FormatLevel     string
	Loglvl          string
}

func init() {
	output := zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatTimestamp: func(i interface{}) string {
			return time.DateTime
		},
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("%-6s", i))
		},
	}

	log.Logger = zerolog.New(output).With().
		Timestamp().CallerWithSkipFrameCount(2).Logger()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func Configure(cfg LoggerCfg) { //при подключении пакета?
	output := zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatTimestamp: func(i interface{}) string {
			parse, _ := time.Parse(time.RFC3339, i.(string))
			return parse.Format(cfg.FormatTimestamp)
		},
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf(cfg.FormatLevel, i))
		},
	}

	log.Logger = zerolog.New(output).With().
		Timestamp().CallerWithSkipFrameCount(2).Logger()

	lvl, err := zerolog.ParseLevel(cfg.Loglvl)
	if err != nil {
		log.Warn().Err(fmt.Errorf("parseLevel: %w", err)).Msg("applogger.Configure")
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
}
