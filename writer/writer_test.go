package writer

import (
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/sysdevguru/bufnet/reader"
)

func TestWrite(t *testing.T) {
	t.Parallel()

	// test writing 1024 * 1024 data with 500 * 1024 buffer
	// expected time is 1.946s~2.151s
	tr := &reader.TestReader{Size: 1 << 20}
	bw := NewWriter(ioutil.Discard, 500<<10)

	start := time.Now()
	n, err := io.Copy(bw, tr)
	dur := time.Since(start)
	if err != nil {
		t.Error(err)
	}
	if n != 1<<20 {
		t.Errorf("Want %d bytes, got %d.", 1<<20, n)
	}
	t.Logf("Wrote %d bytes in %s.", n, dur)
	if dur < 1946*time.Millisecond || dur > 2151*time.Millisecond {
		t.Errorf("Took %s, want 1.946s~2.151s.", dur)
	}
}
