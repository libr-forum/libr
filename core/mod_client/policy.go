package main

import (
	"context"
	"regexp"
	"strings"

	service "github.com/libr-forum/Libr/core/libr/mod/service"
	"github.com/libr-forum/Libr/core/libr/models"
	"github.com/libr-forum/Libr/core/libr/moderator"
	"github.com/libr-forum/Libr/core/libr/proto"
)

var (
	imageRe = regexp.MustCompile(`(?i)<img\s+[^>]*src="data:image/`)
	linkRe  = regexp.MustCompile(`(?i)\b((?:https?://|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)[^\s]+)`)
)

// postPolicy is the v1 moderation pipeline as a v2 policy: images go to the human
// review queue; forbidden words and links are rejected; text is judged by the
// configured LLM, else Google NLP. When the classifier fails the post goes to a human
// (v1 approved it).
func postPolicy(ctx context.Context, m proto.Msg) (bool, error) {
	ok, err := classify(string(m.Content))
	status := "0"
	switch {
	case err == moderator.ErrDefer:
		status = "pending"
	case ok:
		status = "1"
	}
	_ = service.AppendToModLog(models.UserMsg{Content: string(m.Content), TimeStamp: m.Ts}, status)
	return ok, err
}

func classify(content string) (bool, error) {
	if imageRe.MatchString(content) {
		return false, moderator.ErrDefer
	}
	lower := strings.ToLower(content)
	for _, w := range service.LoadForbiddenWords() {
		if w != "" && strings.Contains(lower, strings.ToLower(w)) {
			return false, nil
		}
	}
	if linkRe.MatchString(content) {
		return false, nil
	}
	if cfg, err := service.ReadModConfigFile(); err == nil && cfg.LLMEnabled {
		if clean, err := service.AnalyzeWithLLM(content, cfg); err == nil {
			return clean, nil
		}
	}
	clean, err := service.AnalyzeWithGoogleNLP(content)
	if err != nil {
		return false, moderator.ErrDefer
	}
	return clean, nil
}
