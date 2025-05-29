package logger

import (
    "fmt"
    "log"
    "os"
)

type stdLogger struct {
    debug *log.Logger
    info  *log.Logger
    warn  *log.Logger
    error *log.Logger
    fatal *log.Logger
}

// NewStdLogger создает новый логгер на основе стандартного log пакета
func NewStdLogger() Logger {
    return &stdLogger{
        debug: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
        info:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
        warn:  log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
        error: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
        fatal: log.New(os.Stderr, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile),
    }
}

func (l *stdLogger) Debug(msg string, fields ...Field) {
    l.debug.Output(2, formatMsg(msg, fields...))
}

func (l *stdLogger) Info(msg string, fields ...Field) {
    l.info.Output(2, formatMsg(msg, fields...))
}

func (l *stdLogger) Warn(msg string, fields ...Field) {
    l.warn.Output(2, formatMsg(msg, fields...))
}

func (l *stdLogger) Error(msg string, fields ...Field) {
    l.error.Output(2, formatMsg(msg, fields...))
}

func (l *stdLogger) Fatal(msg string, fields ...Field) {
    l.fatal.Output(2, formatMsg(msg, fields...))
    os.Exit(1)
}

// formatMsg форматирует сообщение с дополнительными полями
func formatMsg(msg string, fields ...Field) string {
    if len(fields) == 0 {
        return msg
    }
    
    result := msg + " {"
    for i, f := range fields {
        if i > 0 {
            result += ", "
        }
        result += fmt.Sprintf("%s: %v", f.Key, f.Value)
    }
    result += "}"
    return result
}