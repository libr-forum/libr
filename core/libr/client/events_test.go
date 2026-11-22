package client

import (
	"testing"

	"github.com/libr-forum/Libr/core/libr/cache"
	"github.com/libr-forum/Libr/core/libr/types"
)

type recorder struct{ finalized []types.FinalizedModeration }

func (r *recorder) CronStatus([]map[string]interface{}) {}
func (r *recorder) ModerationFinalized(f types.FinalizedModeration) {
	r.finalized = append(r.finalized, f)
}

func TestFinalizeEmitsAndRecords(t *testing.T) {
	cache.SetCacheDir(t.TempDir())
	rec := &recorder{}
	SetEvents(rec)
	t.Cleanup(func() { SetEvents(nil) })

	finalize(types.PendingModeration{MsgSign: "sig-img", AckCount: 3, MsgCert: types.MsgCert{Type: "manual_mod"}}, "approved", 2, 1)
	finalize(types.PendingModeration{MsgSign: "sig-rep", AckCount: 1}, "rejected", 0, 2)

	if len(rec.finalized) != 2 {
		t.Fatalf("events = %d, want 2", len(rec.finalized))
	}
	if f := rec.finalized[0]; f.Type != "manual_mod" || f.Status != "approved" || f.Total != 3 {
		t.Fatalf("image item: %+v", f)
	}
	if f := rec.finalized[1]; f.Type != "report" || f.Status != "rejected" || f.Total != 2 {
		t.Fatalf("report item (total must cover all decisions): %+v", f)
	}
	saved, err := cache.GetAllFinalizedModerations()
	if err != nil || len(saved) != 2 {
		t.Fatalf("saved %d records, err %v", len(saved), err)
	}

	SetEvents(nil)
	finalize(types.PendingModeration{MsgSign: "sig-3"}, "approved", 1, 0) // must not panic without a sink
}
