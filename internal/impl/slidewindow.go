package impl

import (
	"github.com/nedcg/ratelimiter/pkg/config"
	"time"
)

type SlidingWindow struct {
	MaxCapacity     int
	CurrentCapacity int
	PrevCapacity    int
	LastTimeWindow  time.Time
	Clock           config.ClockFunc
}

func NewSlideWindow(config config.ConfigParams) SlidingWindow {
	return SlidingWindow{
		MaxCapacity:     config.Tokens,
		CurrentCapacity: config.Tokens,
		PrevCapacity:    0,
		LastTimeWindow:  config.Clock().Round(time.Second),
		Clock:           config.Clock,
	}
}

func (fw SlidingWindow) Allow() bool {
	now := fw.Clock()
	currentTimeWindow := now.Round(time.Second)

	if fw.LastTimeWindow != currentTimeWindow {
		fw.LastTimeWindow = currentTimeWindow
		fw.PrevCapacity = fw.CurrentCapacity
		fw.CurrentCapacity = 0
	}

	progress := int(now.Sub(currentTimeWindow).Milliseconds()) / 10
	prevProgress := fw.PrevCapacity * (100 - progress) / 100 // 0 to 100%

	if prevProgress+fw.CurrentCapacity+1 <= fw.MaxCapacity {
		fw.CurrentCapacity += 1
		return true
	}

	return false
}
