package httpserver

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	_timeout = 15 * time.Second
)

type HttpServer struct {
	httpServe *http.Server
}

func New(baseCtx context.Context, h http.Handler, port int) *HttpServer {
	return &HttpServer{
		httpServe: &http.Server{
			Addr: ":" + strconv.Itoa(port),

			Handler: h,

			WriteTimeout: _timeout,
			ReadTimeout:  _timeout,

			MaxHeaderBytes: 1 << 20,

			BaseContext: func(_ net.Listener) context.Context {
				return baseCtx
			},
		},
	}
}

func (s *HttpServer) Run() error {
	return s.httpServe.ListenAndServe()
}

func (s *HttpServer) ShutDown(ctx context.Context) error {
	return s.httpServe.Shutdown(ctx)
}
