package helper

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ZapLogger wraps a zap.Logger for the lego client.
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new logger.
func NewZapLogger(logger *zap.Logger) *ZapLogger {
	return &ZapLogger{
		logger: logger,
	}
}

// Fatal does logging in fatal level.
func (z *ZapLogger) Fatal(args ...interface{}) {
	z.logger.Sugar().Fatal(args...)
}

// Fatalln is in theory equivalent to Fatal, but followed by a line break. However, for ZAP both functions behave same
func (z *ZapLogger) Fatalln(args ...interface{}) {
	z.Fatal(args...)
}

// Fatalf is equivalent to Fatalln, but with formatting.
func (z *ZapLogger) Fatalf(format string, args ...interface{}) {
	z.logger.Sugar().Fatalf(format, args...)
}

// Print is equivalent to Println, but without the final line break.
func (z *ZapLogger) Print(args ...interface{}) {
	msg := strings.TrimSpace(fmt.Sprint(args...))
	if strings.HasPrefix(msg, "[WARN]") {
		z.logger.Sugar().Warn(strings.TrimPrefix(msg, "[WARN]"))
	} else {
		z.logger.Sugar().Info(msg)
	}
}

// Println is in theory equivalent to Print, but followed by a line break. However, for ZAP both functions behave same
func (z *ZapLogger) Println(args ...interface{}) {
	z.Print(args...)
}

// Printf is equivalent to Print, but with formatting.
func (z *ZapLogger) Printf(format string, args ...interface{}) {
	msg := strings.TrimSpace(fmt.Sprintf(format, args...))
	if strings.HasPrefix(msg, "[WARN]") {
		z.logger.Sugar().Warn(strings.TrimPrefix(msg, "[WARN]"))
	} else {
		z.logger.Sugar().Info(msg)
	}

}
