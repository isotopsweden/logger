package httplogger

import (
	"net/http"
	"regexp"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/segmentio/go-log"
)

// Logger middleware.
type Logger struct {
	h          http.Handler
	log        *log.Logger
	processors map[string]string
}

// SetLogger sets the logger to `log`.
func (l *Logger) SetLogger(log *log.Logger) {
	l.log = log
}

// wrapper to capture status.
type wrapper struct {
	http.ResponseWriter
	written int
	status  int
}

// capture status.
func (w *wrapper) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// capture written bytes.
func (w *wrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.written += n
	return n, err
}

// New logger middleware.
func New(args ...interface{}) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		l := &Logger{
			log: log.Log,
			h:   h,
		}

		if len(args) > 0 {
			if p, ok := args[0].(map[string]string); ok {
				l.processors = p
			}
		}

		return l
	}
}

// NewLogger logger middleware with the given logger.
func NewLogger(log *log.Logger, args ...interface{}) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		l := &Logger{
			log: log,
			h:   h,
		}

		if len(args) > 0 {
			if p, ok := args[0].(map[string]string); ok {
				l.processors = p
			}
		}

		return l
	}
}

// ServeHTTP implementation.
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	res := &wrapper{w, 0, 200}

	l.log.Info(">> %s %s", r.Method, l.Sanitize(r.RequestURI))
	l.h.ServeHTTP(res, r)
	size := humanize.Bytes(uint64(res.written))

	switch {
	case res.status >= 500:
		l.log.Error("<< %s %s %d (%s) in %s", r.Method, l.Sanitize(r.RequestURI), res.status, size, time.Since(start))
	case res.status >= 400:
		l.log.Warning("<< %s %s %d (%s) in %s", r.Method, l.Sanitize(r.RequestURI), res.status, size, time.Since(start))
	default:
		l.log.Info("<< %s %s %d (%s) in %s", r.Method, l.Sanitize(r.RequestURI), res.status, size, time.Since(start))
	}
}

// Sanitize input.
func (l *Logger) Sanitize(s string) string {
	if l.processors == nil {
		return s
	}

	for key, value := range l.processors {
		r, err := regexp.Compile(key)
		if err != nil {
			continue
		}
		s = r.ReplaceAllString(s, value)
	}

	return s
}
