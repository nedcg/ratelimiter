package config

import "time"

type ClockFunc func() time.Time

type ConfigParams struct {
	Tokens     int
	RefillRate int
	Clock      ClockFunc
}
