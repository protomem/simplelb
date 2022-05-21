package loadbalancer

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/protomem/simplelb/pkg/backend"
	"github.com/protomem/simplelb/pkg/backendpool"
	"github.com/sirupsen/logrus"
)

const (
	Attempts KeyCtx = "att"
	Retry    KeyCtx = "ret"
)

type KeyCtx string

type LB struct {
	backPool *backendpool.BP
}

func New(bp *backendpool.BP) *LB {
	return &LB{
		backPool: bp,
	}
}

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

func (lb *LB) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		logrus.Errorf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "service not available", http.StatusServiceUnavailable)
		return
	}

	bakc := lb.backPool.Backend()
	if bakc == nil {
		http.Error(w, "service not available", http.StatusServiceUnavailable)
		return
	}

	logrus.Infof("Get Backend: %s", bakc.URL())

	targetURL, _ := url.Parse(bakc.URL())
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logrus.Infof("[%s] %s\n", targetURL.Host, err.Error())

		retries := GetRetryFromContext(r)
		if retries < 3 {
			<-time.After(10 * time.Millisecond)
			ctx := context.WithValue(r.Context(), Retry, retries+1)
			proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		bakc.SetStatus(backend.StatusNotAvailable)

		attempts := GetAttemptsFromContext(r)
		logrus.Infof("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
		ctx := context.WithValue(r.Context(), Attempts, attempts+1)
		lb.ServeHTTP(w, r.WithContext(ctx))
	}

	proxy.ServeHTTP(w, r)

}
