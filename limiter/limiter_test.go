package limiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type test struct {
	written    int
	bufferSize int
	lim        Limiter
}

func TestLimit(t *testing.T) {
	testObj := test{written: 1024, bufferSize: 4096, lim: Limiter{Bandwidth: 1024}}
	lim := Limiter{Bandwidth: testObj.lim.Bandwidth}
	lim.Init()
	lim.Limit(testObj.written, testObj.bufferSize)
	assert.Equal(t, 0, int(lim.Bucket))

	// bandwidth is negative
	testObj = test{written: 1024, bufferSize: 4096, lim: Limiter{Bandwidth: -1024}}
	lim = Limiter{Bandwidth: testObj.lim.Bandwidth}
	lim.Init()
	lim.Limit(testObj.written, testObj.bufferSize)
	assert.Equal(t, 0, int(lim.Bucket))

	// elapsed time is larger than stall threshold
	testObj = test{written: 2048, bufferSize: 4096, lim: Limiter{Bandwidth: 1024}}
	time.Sleep(time.Duration(6 * time.Second))
	lim = Limiter{Bandwidth: testObj.lim.Bandwidth}
	lim.Init()
	lim.Limit(testObj.written, testObj.bufferSize)
	assert.Equal(t, 0, int(lim.Bucket))
}
