package wire

import (
	"bufio"
	"bytes"
	"testing"
)

func TestBucket(t *testing.T) {
	for _, tc := range []struct{ ts, want int64 }{
		{0, 0}, {899, 0}, {900, 900}, {1_800_000_123, 1_800_000_123 - 1_800_000_123%900},
	} {
		if got := Bucket(tc.ts); got != tc.want {
			t.Errorf("Bucket(%d) = %d, want %d", tc.ts, got, tc.want)
		}
	}
	if IndexKey("a", 0) == IndexKey("a", 900) || IndexKey("a", 0) == IndexKey("b", 0) {
		t.Error("index keys collide across buckets or communities")
	}
}

func TestLimiterPerKey(t *testing.T) {
	l := NewLimiter(0.001, 2)
	got := []bool{l.Allow("a"), l.Allow("a"), l.Allow("a")}
	if !got[0] || !got[1] || got[2] {
		t.Fatalf("burst of 2: got %v", got)
	}
	if !l.Allow("b") {
		t.Fatal("one key's budget spent another's")
	}
}

func TestFrameByteCount(t *testing.T) {
	var buf bytes.Buffer
	n, err := writeFrame(&buf, map[string]string{"k": "v"})
	if err != nil || n != buf.Len() {
		t.Fatalf("wrote %d, buffer %d, err %v", n, buf.Len(), err)
	}
	var v map[string]string
	m, err := readFrame(bufio.NewReader(&buf), &v)
	if err != nil || m != n || v["k"] != "v" {
		t.Fatalf("read %d of %d: %v %v", m, n, v, err)
	}
}
