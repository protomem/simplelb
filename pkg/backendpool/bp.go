package backendpool

import (
	"errors"
	"sync/atomic"

	"github.com/protomem/simplelb/pkg/backend"
)

type BP struct {
	backends []*backend.Backend
	current  uint64
}

func New(addrs ...string) (*BP, error) {
	if len(addrs) == 0 {
		return nil, errors.New("address list is empty")
	}

	backends := make([]*backend.Backend, 0, len(addrs))

	for _, addr := range addrs {
		bakc, err := backend.New(addr)
		if err != nil {
			continue
		}
		backends = append(backends, bakc)
	}

	if len(backends) == 0 {
		return nil, errors.New("no backends available")
	}

	return &BP{
		backends: backends,
		current:  0,
	}, nil
}

// Возвращает индекс следущего backend
func (bp *BP) NextIndex() int {
	return int(atomic.AddUint64(&bp.current, uint64(1)) % uint64(len(bp.backends)))
}

// Возвращает доступный backend
func (bp *BP) Backend() *backend.Backend {
	next := bp.NextIndex()
	l := len(bp.backends) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(bp.backends)                               // take an index by modding
		if bp.backends[idx].Status() == backend.StatusAvailable { // if we have an alive backend, use it and store if its not the original one
			if i != next {
				atomic.StoreUint64(&bp.current, uint64(idx))
			}
			return bp.backends[idx]
		}
	}
	return nil
}
