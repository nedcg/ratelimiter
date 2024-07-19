package impl

import (
	"github.com/nedcg/ratelimiter/pkg/config"
	"time"
)

type FixedWindow struct {
	MaxCapacity     int
	CurrentCapacity int
	LastWindow      time.Time
	Clock           config.ClockFunc
}

func NewFixedWindow(config config.ConfigParams) FixedWindow {
	return FixedWindow{
		MaxCapacity:     config.Tokens,
		CurrentCapacity: config.Tokens,
		LastWindow:      config.Clock().Round(time.Second),
	}
}

func (fw FixedWindow) Allow() bool {
	now := fw.Clock().Round(time.Second)

	if fw.LastWindow == now {
		fw.LastWindow = now
		fw.CurrentCapacity = fw.MaxCapacity
	}

	if fw.CurrentCapacity > 0 {
		fw.CurrentCapacity--
		return true
	}

	return false
}
