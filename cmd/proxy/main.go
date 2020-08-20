package main

import (
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/valyala/fasthttp"

	"github.com/ip-api/proxy/internal/batch"
	"github.com/ip-api/proxy/internal/cache"
	"github.com/ip-api/proxy/internal/fetcher"
	"github.com/ip-api/proxy/internal/handlers"
	"github.com/ip-api/proxy/internal/reverse"
	"github.com/ip-api/proxy/internal/util"
)

func main() {
	var logger zerolog.Logger

	if os.Getenv("LOG_OUTPUT") == "console" {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05.000",
		})
	} else {
		// Use structured JSON logging compatible with Google Stack Driver.

		// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
		zerolog.MessageFieldName = "message"
		zerolog.CallerFieldName = "caller"
		zerolog.ErrorStackFieldName = "stacktrace"
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimestampFieldName = "timestamp"
		zerolog.TimeFieldFormat = time.RFC3339Nano
		zerolog.TimestampFunc = func() time.Time {
			return time.Now().UTC()
		}

		// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
		zerolog.LevelFieldName = "severity"
		zerolog.LevelFieldMarshalFunc = convertLevelToStackdriver

		logger = zerolog.New(os.Stderr)
		logger = logger.Hook(zerolog.HookFunc(defaultLevelToInfo))
	}

	// Default is to log everything.
	switch os.Getenv("LOG_LEVEL") {
	case "info":
		logger = logger.Level(zerolog.InfoLevel)
	case "warn":
		logger = logger.Level(zerolog.WarnLevel)
	case "error":
		logger = logger.Level(zerolog.ErrorLevel)
	}

	logger = logger.With().Str("part", "main").Logger()

	reverser := reverse.New(logger.With().Str("part", "reverser").Logger())

	client, err := fetcher.NewIPApi(logger.With().Str("part", "fetcher").Logger(), reverser)
	if err != nil {
		logger.Fatal().Err(err).Msg("could not create fetcher")
	}

	cacheSize := 1024 * 1024 * 1024 // 1GB
	if v := os.Getenv("CACHE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err != nil {
			logger.Fatal().Err(err).Msg("invalid CACHE_SIZE")
		} else {
			cacheSize = n
		}
	}

	cache := cache.New(cacheSize)
	batches := batch.New(logger.With().Str("part", "batch").Logger(), cache, client)

	go batches.ProcessLoop()

	h := handlers.Handler{
		Logger:  logger.With().Str("part", "handler").Logger(),
		Batches: batches,
		Client:  client,
	}

	s := &fasthttp.Server{
		Handler:               h.Index,
		ReadTimeout:           time.Minute,
		WriteTimeout:          time.Minute,
		IdleTimeout:           time.Hour,
		ReadBufferSize:        4096 * 2,
		WriteBufferSize:       4096 * 2,
		MaxRequestBodySize:    1 * 1024 * 1024,
		Logger:                util.FasthttpLogger{Logger: logger.With().Str("part", "fasthttp").Logger()},
		NoDefaultServerHeader: true,
		NoDefaultContentType:  true,
	}

	addr := os.Getenv("LISTEN")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	logger.Info().Msgf("listening on %q", addr)

	go func() {
		if err := s.ListenAndServe(addr); err != nil {
			logger.Fatal().Err(err).Msg("failed to ListenAndServe")
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-ch
	signal.Stop(ch)

	// Shut down the http server so another process can take over.
	// But wait for 10 seconds to give active connections some time to terminate gracefully.
	go func() {
		if err := s.Shutdown(); err != nil {
			logger.Error().Err(err).Msg("failed to shutdown server")
		}
	}()

	time.Sleep(time.Second * 10)
}

// convertLevelToStackdriver converts a zerolog.Level to a stackdriver compatible
// sererity string.
//
// See: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
func convertLevelToStackdriver(l zerolog.Level) string {
	switch l {
	case zerolog.NoLevel:
		fallthrough
	case zerolog.DebugLevel:
		return "DEBUG"
	case zerolog.InfoLevel:
		return "INFO"
	case zerolog.WarnLevel:
		return "WARNING"
	case zerolog.ErrorLevel:
		return "ERROR"
	case zerolog.FatalLevel:
		return "CRITICAL"
	case zerolog.PanicLevel:
		return "EMERGENCY"
	default:
		return strings.ToUpper(l.String())
	}
}

// defaultLevelToInfo is a zerolog Hook that add severity:INFO  to an event
// if it has the level 'NoLevel' set.
func defaultLevelToInfo(e *zerolog.Event, level zerolog.Level, message string) {
	if level == zerolog.NoLevel {
		e.Str("severity", "INFO")
	}
}
