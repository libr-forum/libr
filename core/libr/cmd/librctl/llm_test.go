package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/libr-forum/Libr/core/libr/moderator"
	"github.com/libr-forum/Libr/core/libr/proto"
)

// fakeOllama answers reject when the text or image matches bad, approve otherwise,
// and records which models it was asked.
func fakeOllama(t *testing.T, bad string, models *[]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Model    string
			Messages []struct {
				Content string
				Images  []string
			}
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		*models = append(*models, req.Model)
		u := req.Messages[1]
		d := "approve"
		if strings.Contains(u.Content, bad) || (len(u.Images) > 0 && u.Images[0] == bad) {
			d = "reject"
		}
		if strings.Contains(u.Content, "garble") {
			_ = json.NewEncoder(w).Encode(map[string]any{"message": map[string]string{"content": "not json"}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"message": map[string]string{"content": `{"decision":"` + d + `","reason":"x"}`}})
	}))
}

func TestLLMPolicy(t *testing.T) {
	var models []string
	srv := fakeOllama(t, "BAD", &models)
	defer srv.Close()
	c := llmConfig{URL: srv.URL, Model: "text", VisionModel: "vision", Rules: map[string]string{"g": "no BAD"}}
	pol := c.policy(srv.Client())
	ask := func(community, content string) (bool, error) {
		return pol(context.Background(), proto.Msg{Community: community, Content: []byte(content)})
	}
	if ok, err := ask("g", "hello"); !ok || err != nil {
		t.Fatalf("clean text: %v %v", ok, err)
	}
	if ok, err := ask("g", "BAD words"); ok || err != nil {
		t.Fatalf("bad text: %v %v", ok, err)
	}
	if ok, err := ask("g", `look <img src="data:image/png;base64,BAD" />`); ok || err != nil {
		t.Fatalf("bad image: %v %v", ok, err)
	}
	models = nil
	if ok, err := ask("g", `<img src="data:image/jpeg;base64,AAAA" />`); !ok || err != nil || len(models) != 1 || models[0] != "vision" {
		t.Fatalf("image-only post: %v %v %v", ok, err, models)
	}
	for _, tc := range []struct{ community, content string }{
		{"other", "hello"},                      // no rules
		{"g", "garble"},                         // unreadable answer
		{"g", `<img src="data:image/png,raw">`}, // image the policy cannot extract
	} {
		if ok, err := ask(tc.community, tc.content); ok || !errors.Is(err, moderator.ErrDefer) {
			t.Fatalf("%q in %q: want deferral, got %v %v", tc.content, tc.community, ok, err)
		}
	}
	c.VisionModel = ""
	if ok, err := c.policy(srv.Client())(context.Background(), proto.Msg{Community: "g", Content: []byte(`<img src="data:image/png;base64,AAAA">`)}); ok || !errors.Is(err, moderator.ErrDefer) {
		t.Fatalf("no vision model: want deferral, got %v %v", ok, err)
	}
	srv.Close()
	if ok, err := ask("g", "hello"); ok || !errors.Is(err, moderator.ErrDefer) {
		t.Fatalf("server down: want deferral, got %v %v", ok, err)
	}
}
