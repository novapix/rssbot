package logger

import (
	"io"
	"log"
	"os"
)

var (
	Info  *log.Logger
	Error *log.Logger
)

func Init(logFile string) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file %s: %v", logFile, err)
	}

	multiOut := io.MultiWriter(file, os.Stdout)

	Info = log.New(multiOut, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(multiOut, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
