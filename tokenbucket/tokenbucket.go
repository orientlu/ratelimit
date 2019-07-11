package tokenbucket

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// TokenBucket ...
type TokenBucket struct {
	ratePerSecond               int
	updatePeriod                time.Duration
	updateSlotNumber            int
	putTokenNumberPerSlot       int
	remaindTokenNumberPerSecond int
	slotCount                   int
	ch                          chan struct{}
	quitFun                     context.CancelFunc
}

// New create a TokenBucket an return
func New(ratePerSecond int, capacity int, updatePeriod time.Duration) (*TokenBucket, error) {
	if ratePerSecond <= 0 || capacity <= 0 || updatePeriod > time.Second {
		return nil, errors.New("new tokenbucket error")
	}
	t := &TokenBucket{
		ratePerSecond: ratePerSecond,
		updatePeriod:  updatePeriod,
	}
	t.updateSlotNumber = int(time.Second / t.updatePeriod)
	t.putTokenNumberPerSlot = ratePerSecond / t.updateSlotNumber
	t.ch = make(chan struct{}, capacity)
	return t, nil
}

// Update is a loop, will quit until call StopUpdate
// add the token into bucket at the rate
func (t *TokenBucket) Update() {
	ticker := time.NewTicker(t.updatePeriod)
	t.slotCount = 0
	t.remaindTokenNumberPerSecond = t.ratePerSecond % t.updateSlotNumber

	var ctx context.Context
	ctx, t.quitFun = context.WithCancel(context.Background())
	for {
		select {
		case <-ticker.C:
			putTokenNumber := t.putTokenNumberPerSlot
			if t.remaindTokenNumberPerSecond > 0 {
				putTokenNumber++
				t.remaindTokenNumberPerSecond--
			}
			for i := 0; i < putTokenNumber; i++ {
				select {
				case t.ch <- struct{}{}:
				default:
					// full
					break
				}
			}

			t.slotCount++
			if t.slotCount == t.updateSlotNumber {
				t.slotCount = 0
				t.remaindTokenNumberPerSecond = t.ratePerSecond % t.updateSlotNumber
			}

		case <-ctx.Done():
			return
		}
	}
}

// StopUpdate let Update loop stop
func (t *TokenBucket) StopUpdate() error {
	if t.quitFun == nil {
		return errors.New("stop tokenbucket error,no update loop run")
	}
	t.quitFun()
	t.quitFun = nil
	return nil
}

// TryAcquire will not block
func (t *TokenBucket) TryAcquire() bool {
	select {
	case <-t.ch:
		return true
	default:
		return false
	}
}

// Acquire try to get a token, or block to wait
func (t *TokenBucket) Acquire() bool {
	<-t.ch
	return true
}

// GetInfo return the capacity, tokenCount, ratePerSecond
func (t *TokenBucket) GetInfo() (int, int, int) {
	return cap(t.ch), len(t.ch), t.ratePerSecond
}
