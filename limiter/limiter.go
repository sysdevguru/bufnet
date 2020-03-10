package limiter

import (
	"fmt"
	"time"
)

var (
	thresholdRatio = 0.05
)

// Limiter limits bandwidth
type Limiter struct {
	Bandwidth   int
	Bucket      int64
	Initialized bool
	Start       time.Time
}

// Init initialize Limiter
func (l *Limiter) Init() {
	if !l.Initialized {
		l.reset()
		l.Initialized = true
	}
}

func (l *Limiter) reset() {
	l.Bucket = 0
	l.Start = time.Now()
}

// Limit is the function that actually limits bandwidth
func (l *Limiter) Limit(n, bufSize int) {
	// not apply limit in case desired bandwidth is 0 or negative
	if l.Bandwidth <= 0 {
		return
	}

	l.Bucket += int64(n)

	// elapsed time for the read/write operation
	elapsed := time.Since(l.Start)
	// sleep for the keeped time and reset limiter
	keepedTime := time.Duration(l.Bucket)*time.Second/time.Duration(l.Bandwidth) - elapsed
	if keepedTime > 0 {
		time.Sleep(keepedTime)
		l.reset()
		return
	}

	// reset the limiter when elapsed time is out of thresholds
	// current threshold is estimated time +/- 5%
	estimation := time.Duration(bufSize/l.Bandwidth) * time.Second
	upperThreshold := time.Duration(thresholdRatio)*estimation + estimation
	lowerThreshold := estimation - estimation*time.Duration(thresholdRatio)
	if elapsed > upperThreshold {
		l.reset()
		return
	}
	if elapsed < lowerThreshold {
		fmt.Println("here")
		time.Sleep(elapsed - lowerThreshold)
		l.reset()
	}
}
