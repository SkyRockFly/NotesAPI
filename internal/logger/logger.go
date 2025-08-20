package applogger

import (
	appconfig "NotesService/configs"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Configure(cfg *appconfig.Config) {
	output := zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatTimestamp: func(i interface{}) string {
			parse, _ := time.Parse(time.RFC3339, i.(string))
			return parse.Format(cfg.Log.FormatLevel)
		},
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf(cfg.Log.FormatLevel, i))
		},
	}

	log.Logger = zerolog.New(output).With().
		Timestamp().CallerWithSkipFrameCount(2).Logger()

	lvl, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
}
