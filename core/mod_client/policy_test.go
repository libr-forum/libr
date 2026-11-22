package main

import (
	"testing"

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
