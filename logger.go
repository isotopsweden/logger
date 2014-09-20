package logger

import "github.com/dustin/go-humanize"
import "github.com/segmentio/go-log"
import "net/http"
import "time"

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
func New() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			res := &wrapper{w, 0, 200}
			log.Info(">> %s %s", req.Method, req.RequestURI)
			h.ServeHTTP(res, req)
			size := humanize.Bytes(uint64(res.written))
			log.Info("<< %s %s %d (%s) in %s", req.Method, req.RequestURI, res.status, size, time.Since(start))
		})
	}
}