package zergrepo

import (
	"runtime"
	"strings"
)

// WarnIfFail logs a function error if an error occurs.
func (r *Repo) WarnIfFail(fn func() error) {
	if err := fn(); err != nil {
		r.log.Warnf("failed call func: %s", err)
	}
}

func callerMethodName() string {
	pc, _, _, _ := runtime.Caller(2)
	names := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	return names[len(names)-1]
}
