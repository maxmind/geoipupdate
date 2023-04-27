package vars

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// NewDiscardLogger returns a logger with an io.Discard writer that should
// be replaced with another writer according to need.
// It takes a string reference as an argument to be used as the prefix for
// all log entries.
func NewDiscardLogger(s string) *log.Logger {
	return log.New(io.Discard, prefix(s), log.LstdFlags)
}

// NewBareDiscardLogger returns a logger with an io.Discard writer that should
// be replaced with another writer according to need. It's mainly used to output
// messages without any formatting.
func NewBareDiscardLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

// NewBareStderrLogger returns a bare stderr logger mainly used to output messages
// without any formatting.
func NewBareStderrLogger() *log.Logger {
	return log.New(os.Stderr, "", 0)
}

// prefix transforms a string into a consistent prefix of a certain
// size and form to be used in various loggers.
func prefix(s string) string {
	size := 6
	format := "[%s] "

	if len(s) > size {
		s = s[:size]
	}

	if len(s) < size {
		spaces := strings.Repeat(" ", size-len(s))
		s += spaces
	}

	s = strings.ToUpper(s)
	return fmt.Sprintf(format, s)
}
