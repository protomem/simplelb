package backend

import (
	"fmt"
	"net/url"
	"sync"
)

const (
	_ Status = iota
	StatusAvailable
	StatusNotAvailable
)

type Status int

func (s Status) String() string {
	if s == StatusAvailable {
		return "available"
	} else {
		return "not available"
	}
}

type Backend struct {
	mux sync.RWMutex

	url    *url.URL
	status Status
}

func New(rawURL string) (*Backend, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &Backend{
		url:    url,
		status: StatusAvailable,
	}, nil
}

func (b *Backend) Status() Status {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return b.status
}

func (b *Backend) SetStatus(s Status) {
	b.mux.Lock()
	defer b.mux.Unlock()

	b.status = s
}

func (b *Backend) URL() string {
	b.mux.RLock()
	defer b.mux.RUnlock()

	return b.url.String()
}

func (b *Backend) String() string {
	return fmt.Sprintf("%s: %s", b.URL(), b.status)
}
