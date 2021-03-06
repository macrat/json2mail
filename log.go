package main

import (
	"encoding/json"
	"io"
	"time"
)

type Logger struct {
	encoder *json.Encoder
}

func NewLogger(w io.Writer) *Logger {
	return &Logger{json.NewEncoder(w)}
}

func (l *Logger) write(msg interface{}) {
	l.encoder.Encode(msg)
}

func getTime() string {
	return time.Now().Format(time.RFC3339)
}

type errorLog struct {
	Timestamp string      `json:"timestamp"`
	Error     string      `json:"error"`
	Input     interface{} `json:"input_data"`
}

func (l *Logger) Error(msg string, input interface{}) {
	l.write(errorLog{
		Timestamp: getTime(),
		Error:     msg,
		Input:     input,
	})
}

type mailLog struct {
	Timestamp string `json:"timestamp"`
	Mail      Mail   `json:"mail"`
}

func (l *Logger) Mail(m Mail) {
	l.write(mailLog{
		Timestamp: getTime(),
		Mail:      m,
	})
}
