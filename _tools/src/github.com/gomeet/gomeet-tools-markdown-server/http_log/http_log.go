package http_log

import (
	"log"
	"net/http"
)

type loggingResponseWriter struct {
	writer http.ResponseWriter
	status int
}

func (w *loggingResponseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	w.status = http.StatusOK
	return w.writer.Write(b)
}

func (w *loggingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.writer.WriteHeader(status)
}

type requestLogger struct {
	Handler http.Handler
}

func Log(handler http.Handler) http.Handler {
	return &requestLogger{handler}
}

func (l requestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lw := &loggingResponseWriter{writer: w}
	l.Handler.ServeHTTP(lw, r)
	log.Printf(`{ "remote_addr":"%v","request_method":"%v","request_uri":"%v","request_proto":"%v","status":%v}`, r.RemoteAddr, r.Method, r.RequestURI, r.Proto, lw.status)
}
