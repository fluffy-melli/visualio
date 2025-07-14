package log

import (
	"log"
	"os"
)

type Logger struct {
	Error       *os.File
	ErrorLogger *log.Logger
}

func NewLogger(errorFilePath string) *Logger {
	errorFile, err := os.OpenFile(errorFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	errorLogger := log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	return &Logger{
		Error:       errorFile,
		ErrorLogger: errorLogger,
	}
}

func (l *Logger) Printf(format string, v ...any) {
	l.ErrorLogger.Printf(format, v...)
}

func (l *Logger) Println(v ...any) {
	l.ErrorLogger.Println(v...)
}

func (l *Logger) Print(v ...any) {
	l.ErrorLogger.Print(v...)
}

func (l *Logger) Fatal(v ...any) {
	l.ErrorLogger.Fatal(v...)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.ErrorLogger.Fatalf(format, v...)
}

func (l *Logger) Panic(v ...any) {
	l.ErrorLogger.Panic(v...)
}

func (l *Logger) Panicf(format string, v ...any) {
	l.ErrorLogger.Panicf(format, v...)
}

func (l *Logger) Close() error {
	return l.Error.Close()
}
