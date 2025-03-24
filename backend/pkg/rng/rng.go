package rng

import (
	"math/rand"
	"slices"
	"sync"
	"time"
)

const (
	numberOfPicks = 20
	maxNumber     = 80
	drawDelay     = 1 * time.Second
)

type RNGService struct {
	picks []uint8
	mu    sync.RWMutex
}

func NewRNGService() *RNGService {
	rng := &RNGService{
		mu:    sync.RWMutex{},
		picks: make([]uint8, numberOfPicks),
	}
	go rng.drawLoop()
	time.Sleep(5 * time.Millisecond)
	return rng
}

func (r *RNGService) drawLoop() {
	buf := make([]uint8, numberOfPicks)
	for {
		for idx := range numberOfPicks {
			for {
				pick := uint8(rand.Intn(maxNumber) + 1)
				if !slices.Contains(r.picks[:idx], pick) {
					buf[idx] = pick
					break
				}
			}
		}

		r.mu.Lock()
		copy(r.picks, buf)
		r.mu.Unlock()
	}
}

func (r *RNGService) GetDraw() []uint8 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	buf := make([]uint8, numberOfPicks)
	copy(buf, r.picks)

	return buf
}
