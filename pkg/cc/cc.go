package cc

import (
	"context"
	"log" // lets go for standard
	"os"

	"github.com/alecthomas/kong"
)

// Define a custom Counter type to count repeated flags (e.g., -v, -vv, -vvv).
type Counter uint

const noflags = 0

type CommonCtx struct {
	Context  context.Context
	LogLevel LogLevel
	Log      *log.Logger
	Clean    bool
	Noop     bool
	Verbose  bool
}

func NewCommonContext(verbosity Counter, clean bool, noop bool) *CommonCtx {
	level := GetLogLevel(verbosity)

	return &CommonCtx{
		Context:  context.Background(),
		LogLevel: level,
		Log:      log.New(os.Stdout, "", noflags),
		Clean:    clean,
		Noop:     noop,
	}
}

type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace
)

var LogLevelStrings = map[LogLevel]string{
	LogLevelError: "ERROR",
	LogLevelInfo:  "INFO",
	LogLevelDebug: "DEBUG",
	LogLevelTrace: "TRACE",
}

// Map verbosity level (-v, -vv, -vvv) to log levels.
func GetLogLevel(verbosity Counter) LogLevel {
	switch verbosity {
	case 1:
		return LogLevelInfo
	case 2:
		return LogLevelDebug
	case 3:
		return LogLevelTrace
	default:
		return LogLevelError // Default level when no verbosity is set.
	}
}

func (c *CommonCtx) LogError(msg string) {
	if c.LogLevel >= LogLevelError {
		c.Log.Println(msg)

	}
}

func (c *CommonCtx) LogInfo(msg string) {
	if c.LogLevel >= LogLevelInfo {
		c.Log.Println(msg)

	}
}

func (c *CommonCtx) LogDebug(msg string) {
	if c.LogLevel >= LogLevelDebug {
		c.Log.Println(msg)

	}
}

func (c *CommonCtx) LogTrace(msg string) {
	if c.LogLevel >= LogLevelTrace {
		c.Log.Println(msg)
	}
}

func (c *Counter) Decode(ctx *kong.DecodeContext) error {
	*c++
	return nil
}
