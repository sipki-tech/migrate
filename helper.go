package zergrepo

import (
	"go.uber.org/zap"
	"runtime"
	"strings"
)

// WarnIfFail logs a function error if an error occurs.
func (r *Repo) WarnIfFail(fn func() error, fields ...zap.Field) {
	if err := fn(); err != nil {
		r.log.Warn("failed call func", append(fields, zap.Error(err))...)
	}
}

func callerMethodName() string {
	pc, _, _, _ := runtime.Caller(2)
	names := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	return names[len(names)-1]
}
