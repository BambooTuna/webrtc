package publish

import "time"

type Audio interface {
	Next() (data []byte, sampleCount float64, err error)
	Reset() error
	Close() error

	GetSampleRate() float64
	GetLastGranule() uint64
	GetSleepDuration(sampleCount float64) time.Duration
}
