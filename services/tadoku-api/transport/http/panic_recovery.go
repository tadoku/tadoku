package http

import (
	"log/slog"
	stdhttp "net/http"
	"runtime/debug"
)

func withPanicRecovery(logger *slog.Logger, next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		recorder := &statusRecorder{ResponseWriter: w}
		defer func() {
			if value := recover(); value != nil {
				if value == stdhttp.ErrAbortHandler {
					panic(value)
				}
				logger.ErrorContext(r.Context(), "panic recovered", "panic", value, "stack", string(debug.Stack()))
				if recorder.status == 0 {
					recorder.WriteHeader(stdhttp.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}
