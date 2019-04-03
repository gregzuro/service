package log

import (
	"path"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// ContextHook logs context.
// Usage:
//   log.AddHook(ContextHook{})
// Source: https://github.com/Sirupsen/logrus/issues/63
// Watch that space for improvements.
type ContextHook struct{}

func (hook ContextHook) Levels() []log.Level {
	return log.AllLevels
}

func (hook ContextHook) Fire(entry *log.Entry) error {
	pc := make([]uintptr, 4, 4)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()
		if !strings.Contains(name, "github.com/Sirupsen/logrus") {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data["file"] = path.Base(file)
			entry.Data["func"] = path.Base(name)
			entry.Data["line"] = line
			entry.Data["pkg"] = path.Base(path.Dir(file))
			break
		}
	}
	return nil
}
