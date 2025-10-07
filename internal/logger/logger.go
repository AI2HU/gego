package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// LogLevel represents the different logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a configurable logger instance
type Logger struct {
	level    LogLevel
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
}

// Global logger instance
var globalLogger *Logger

// Init initializes the global logger with the specified level and output
func Init(level LogLevel, output io.Writer) {
	if output == nil {
		output = os.Stdout
	}

	globalLogger = &Logger{
		level:    level,
		debugLog: log.New(output, fmt.Sprintf("[%s] ", DEBUG.String()), log.LstdFlags),
		infoLog:  log.New(output, fmt.Sprintf("[%s] ", INFO.String()), log.LstdFlags),
		warnLog:  log.New(output, fmt.Sprintf("[%s] ", WARNING.String()), log.LstdFlags),
		errorLog: log.New(output, fmt.Sprintf("[%s] ", ERROR.String()), log.LstdFlags),
	}
}

// ParseLogLevel parses a string log level and returns the corresponding LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING", "WARN":
		return WARNING
	case "ERROR":
		return ERROR
	default:
		return INFO // Default to INFO level
	}
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		// Initialize with default INFO level if not initialized
		Init(INFO, os.Stdout)
	}
	return globalLogger
}

// SetLevel changes the log level of the global logger
func SetLevel(level LogLevel) {
	if globalLogger != nil {
		globalLogger.level = level
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.debugLog.Printf(format, v...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.infoLog.Printf(format, v...)
	}
}

// Warning logs a warning message
func (l *Logger) Warning(format string, v ...interface{}) {
	if l.level <= WARNING {
		l.warnLog.Printf(format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.errorLog.Printf(format, v...)
	}
}

// Fatal logs an error message and exits the program
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.errorLog.Printf(format, v...)
	os.Exit(1)
}

// Global convenience functions
func Debug(format string, v ...interface{}) {
	GetLogger().Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	GetLogger().Info(format, v...)
}

func Warning(format string, v ...interface{}) {
	GetLogger().Warning(format, v...)
}

func Error(format string, v ...interface{}) {
	GetLogger().Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	GetLogger().Fatal(format, v...)
}

// SetOutput changes the output destination for all loggers
func SetOutput(output io.Writer) {
	if globalLogger != nil {
		globalLogger.debugLog.SetOutput(output)
		globalLogger.infoLog.SetOutput(output)
		globalLogger.warnLog.SetOutput(output)
		globalLogger.errorLog.SetOutput(output)
	}
}

// SetFlags changes the flags for all loggers
func SetFlags(flag int) {
	if globalLogger != nil {
		globalLogger.debugLog.SetFlags(flag)
		globalLogger.infoLog.SetFlags(flag)
		globalLogger.warnLog.SetFlags(flag)
		globalLogger.errorLog.SetFlags(flag)
	}
}

// GetLevel returns the current log level
func GetLevel() LogLevel {
	if globalLogger != nil {
		return globalLogger.level
	}
	return INFO
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return GetLevel() <= DEBUG
}

// IsInfoEnabled returns true if info logging is enabled
func IsInfoEnabled() bool {
	return GetLevel() <= INFO
}

// IsWarningEnabled returns true if warning logging is enabled
func IsWarningEnabled() bool {
	return GetLevel() <= WARNING
}

// IsErrorEnabled returns true if error logging is enabled
func IsErrorEnabled() bool {
	return GetLevel() <= ERROR
}
