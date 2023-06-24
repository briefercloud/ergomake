package logger

import (
	"context"
	"os"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var logger zerolog.Logger

func Setup() *zerolog.Logger {
	strLevel := os.Getenv("LOG_LEVEL")
	logLevel, err := zerolog.ParseLevel(strLevel)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger = zerolog.New(os.Stdout).
		Level(logLevel).
		With().
		Timestamp().
		Logger()

	if err != nil {
		logger.Err(err).Str("givenLevel", strLevel).Msg("fail to parse LOG_LEVEL, defaulted to info")
	}

	return &logger
}

func Ctx(ctx context.Context) *zerolog.Logger {
	ctxLogger := log.Ctx(ctx)

	// if logger not present in context, log.Ctx returns a disabled log
	if ctxLogger.GetLevel() == zerolog.Disabled {
		return Get()
	}

	return ctxLogger
}

func Get() *zerolog.Logger {
	return &logger
}

func Middleware(engine *gin.Engine) {
	engine.Use(
		requestid.New(),
		func(c *gin.Context) {
			l := *Get()
			l = l.With().
				Str("request_id", requestid.Get(c)).
				Logger()

			c.Request = c.Request.WithContext(l.WithContext(c.Request.Context()))

			req := c.Request
			res := c.Writer

			logger := logger.With().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Str("ip", c.ClientIP()).
				Logger()

			// Log request
			logger.Info().Msg("Received request")

			// Serve request
			c.Next()

			// Log response
			logger.Info().
				Int("status", res.Status()).
				Msg("Sent response")
		},
	)
}

func With(log *zerolog.Logger) zerolog.Context {
	l := *log
	return l.With()
}
