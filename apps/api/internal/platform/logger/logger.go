package logger

import (
	"io"
	"log/slog"
	"os"
)

// bikin logger yang menulis ke standard output.
func New() *slog.Logger {
	return NewWithWriter(os.Stdout)
}

// bikin logger JSON ke writer yang dikriim.
func NewWithWriter(w io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(w, nil)
	logger := slog.New(handler)
	return logger.With("service", "api")
}
