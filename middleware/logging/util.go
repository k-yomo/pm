package pm_logging

import "time"

func DurationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Microseconds()) / 1000
}
