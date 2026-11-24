package main

import (
	"context"
	"testing"
	"time"

	"github.com/libr-forum/Libr/core/libr/moderator"
)

func TestClassify(t *testing.T) {
	t.Setenv("HOME", t.TempDir()) // no API keys: the classifier is unavailable
	for _, tc := range []struct {
		content string
		ok      bool
		defer_  bool
	}{
		{`<img src="data:image/png;base64,AAAA">`, false, true},
		{"see https://spam.example/x", false, false},
		{"an ordinary sentence", false, true}, // classifier down: a human decides, not auto-approve
	} {
		ok, err := classify(tc.content)
		if ok != tc.ok || (err == moderator.ErrDefer) != tc.defer_ {
			t.Errorf("%q: ok=%v err=%v", tc.content, ok, err)
		}
	}
}

// TestClassifyTimeout: a hung classifier defers to a human once the decide deadline
// passes, instead of holding the moderator's decision slot.
func TestClassifyTimeout(t *testing.T) {
	hang := make(chan struct{})
	defer close(hang)
	classifier = func(string) (bool, error) { <-hang; return true, nil }
	defer func() { classifier = classify }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := classifyCtx(ctx, "hello"); err != moderator.ErrDefer {
		t.Fatalf("err = %v, want ErrDefer", err)
	}
}
