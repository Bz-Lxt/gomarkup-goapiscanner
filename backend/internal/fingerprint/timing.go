package fingerprint

import "time"

const BlindThreshold = 2500 * time.Millisecond

func IsTimingAnomaly(d time.Duration) bool {
	return d >= BlindThreshold
}
