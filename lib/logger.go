package constago

import (
	"fmt"
	"log"
	"os"
)

type runLogger struct {
	level int
	l     *log.Logger
}

const (
	stepPrefix   = "== "
	detailIndent = "    "
)

func newRunLogger(cfg *Config) *runLogger {
	level := 1
	if cfg != nil && cfg.Verbose != nil {
		level = *cfg.Verbose
	}
	return &runLogger{
		level: level,
		// No timestamps/prefix by default; keep output stable and readable.
		l: log.New(os.Stderr, "", 0),
	}
}

func (rl *runLogger) enabled(minLevel int) bool {
	return rl != nil && rl.level >= minLevel
}

func (rl *runLogger) Start() {
	if !rl.enabled(1) {
		return
	}
	rl.l.Println("---------------------")
	rl.l.Println("Running Constago...🚀")
}

func (rl *runLogger) Step(format string, args ...any) {
	if rl.enabled(1) {
		msg := fmt.Sprintf(format, args...)
		rl.l.Printf("%s%s", stepPrefix, msg)
	}
}

func (rl *runLogger) Detail(format string, args ...any) {
	if rl.enabled(2) {
		msg := fmt.Sprintf(format, args...)
		rl.l.Printf("%s%s", detailIndent, msg)
	}
}
