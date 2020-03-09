package reader

import (
	"io"
	"io/ioutil"
	"testing"
	"time"
)

func TestRest(t *testing.T) {
	t.Parallel()

	// test reading 2 * 1024 * 1024 data with 500 * 1024 buffer
	// expected time is 5s
	tr := &TestReader{Size: 2 << 20, Stall: 1 * time.Second}
	br := NewReader(tr, 500<<10)

	start := time.Now()
	n, err := io.Copy(ioutil.Discard, br)
	dur := time.Since(start)
	if err != nil {
		t.Error(err)
	}
	if n != 2<<20 {
		t.Errorf("Want %d bytes, got %d.", 2<<20, n)
	}
	t.Logf("Read %d bytes in %s", n, dur)
	if dur < 4600*time.Millisecond || dur > 5400*time.Millisecond {
		t.Errorf("Took %s, want 5s.", dur)
	}
}
