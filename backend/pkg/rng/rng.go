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
		// For each position in the draw...
		for idx := range numberOfPicks {
			// Keep trying until we get a number not already in buf[0:idx].
			for {
				pick := uint8(rand.Intn(maxNumber) + 1)
				if !slices.Contains(buf[:idx], pick) {
					buf[idx] = pick
					break
				}
			}
		}

		// Safely update the stored picks.
		r.mu.Lock()
		copy(r.picks, buf)
		r.mu.Unlock()

		// Sleep briefly before generating a new draw.
		// Adjust the sleep time as needed.
		time.Sleep(drawDelay)
	}
}

func (r *RNGService) GetDraw() []uint8 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	buf := make([]uint8, numberOfPicks)
	copy(buf, r.picks)

	return buf
}
