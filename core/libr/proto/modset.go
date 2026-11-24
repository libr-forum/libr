package proto

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ErrHandoffNotImplemented is returned for any epoch after genesis. Epoch e+1 will be
// signed by 2f+1 of epoch e (or a voter override) once governance is built; until
// then later epochs are refused, so verification fails closed.
var ErrHandoffNotImplemented = errors.New("modset: epoch hand-off not implemented")

// ModEntry is one moderator: their signing key and the libp2p peer ID derived from it,
// so clients can find them on the DHT.
type ModEntry struct {
	PubKey ed25519.PublicKey `json:"pubkey"`
	PeerID string            `json:"peer_id"`
}

// Sig is one signature over a ModSet.
type Sig struct {
	Signer ed25519.PublicKey `json:"signer"`
	Sig    []byte            `json:"sig"`
}

// ModSet is a community's moderator committee for one epoch. Up to F members may be
// faulty; a decision needs Threshold = 2F+1 distinct member votes, which requires at
// least 3F+1 members so two conflicting decisions cannot both reach quorum.
type ModSet struct {
	Community string     `json:"community"`
	Epoch     uint64     `json:"epoch"`
	Mods      []ModEntry `json:"mods"`
	F         uint64     `json:"f"`
	Threshold uint64     `json:"threshold"`
	PrevHash  Hash       `json:"prev_hash"`  // zero at genesis
	TallyHash Hash       `json:"tally_hash"` // election result this set came from; zero at genesis
	Sigs      []Sig      `json:"sigs"`

	members map[string]bool
}

// Bytes is what founders (genesis) or the previous committee sign.
func (s *ModSet) Bytes() []byte {
	mods := append([]ModEntry(nil), s.Mods...)
	sort.Slice(mods, func(i, j int) bool { return bytes.Compare(mods[i].PubKey, mods[j].PubKey) < 0 })
	fields := [][]byte{[]byte(s.Community), U64(s.Epoch), U64(s.F), U64(s.Threshold), s.PrevHash[:], s.TallyHash[:]}
	for _, m := range mods {
		fields = append(fields, m.PubKey, []byte(m.PeerID))
	}
	return Encode(DomainModSet, fields...)
}

// Hash links the next epoch's PrevHash to this set.
func (s *ModSet) Hash() Hash { return Sum(s.Bytes()) }

// IsMember reports whether k is in the set. Valid only after verification.
func (s *ModSet) IsMember(k ed25519.PublicKey) bool { return s.members[string(k)] }

// SignModSet appends priv's signature.
func SignModSet(priv ed25519.PrivateKey, s *ModSet) {
	s.Sigs = append(s.Sigs, Sig{Signer: pub(priv), Sig: ed25519.Sign(priv, s.Bytes())})
}

// PeerIDFor derives the libp2p peer ID for an Ed25519 app key (see peers.StableIdentity).
func PeerIDFor(k ed25519.PublicKey) (string, error) {
	lk, err := crypto.UnmarshalEd25519PublicKey(k)
	if err != nil {
		return "", err
	}
	id, err := peer.IDFromPublicKey(lk)
	return id.String(), err
}

// VerifyModSetChain checks a community's chain from genesis and returns its latest set.
// Genesis must be signed by exactly the trusted founders.
func VerifyModSetChain(founders []ed25519.PublicKey, chain []ModSet) (*ModSet, error) {
	if len(chain) == 0 {
		return nil, errors.New("modset: empty chain")
	}
	g := &chain[0]
	if g.Epoch != 0 || !g.PrevHash.IsZero() || !g.TallyHash.IsZero() {
		return nil, errors.New("modset: genesis must be epoch 0 with zero prev/tally hash")
	}
	if err := g.checkShape(); err != nil {
		return nil, err
	}
	if err := checkSigners(g, founders); err != nil {
		return nil, fmt.Errorf("modset %s genesis: %w", g.Community, err)
	}
	if len(chain) > 1 {
		return nil, ErrHandoffNotImplemented
	}
	return g, nil
}

// checkShape enforces the BFT sizing and key/peer-ID consistency, and indexes members.
func (s *ModSet) checkShape() error {
	if s.Community == "" {
		return errors.New("modset: empty community")
	}
	// Exactly 3f+1: with more members, 2f+1 approvals no longer need an honest
	// majority of the committee's non-faulty members, and f+1 honest rejections no
	// longer make approval impossible.
	if uint64(len(s.Mods)) != 3*s.F+1 {
		return fmt.Errorf("modset %s: %d mods for f=%d, want 3f+1 = %d", s.Community, len(s.Mods), s.F, 3*s.F+1)
	}
	if s.Threshold != 2*s.F+1 {
		return fmt.Errorf("modset %s: threshold %d, want 2f+1 = %d", s.Community, s.Threshold, 2*s.F+1)
	}
	s.members = make(map[string]bool, len(s.Mods))
	for _, m := range s.Mods {
		if len(m.PubKey) != ed25519.PublicKeySize {
			return fmt.Errorf("modset %s: bad key length %d", s.Community, len(m.PubKey))
		}
		if s.members[string(m.PubKey)] {
			return fmt.Errorf("modset %s: duplicate mod key", s.Community)
		}
		s.members[string(m.PubKey)] = true
		if want, err := PeerIDFor(m.PubKey); err != nil || m.PeerID != want {
			return fmt.Errorf("modset %s: peer_id %q does not match key (want %s)", s.Community, m.PeerID, want)
		}
	}
	return nil
}

// checkSigners requires one valid signature from each signer and nothing else.
func checkSigners(s *ModSet, signers []ed25519.PublicKey) error {
	if len(signers) == 0 {
		return errors.New("no trusted signers")
	}
	want := map[string]bool{}
	for _, k := range signers {
		want[string(k)] = true
	}
	got := map[string]bool{}
	body := s.Bytes()
	for _, sg := range s.Sigs {
		k := string(sg.Signer)
		if !want[k] || got[k] {
			return errors.New("unexpected or duplicate signer")
		}
		if !verify(sg.Signer, body, sg.Sig) {
			return errors.New("bad signature")
		}
		got[k] = true
	}
	if len(got) != len(want) {
		return fmt.Errorf("signed by %d of %d founders", len(got), len(want))
	}
	return nil
}

// CommunityFile is the on-disk/shipped form of one community's chain.
type CommunityFile struct {
	Founders []ed25519.PublicKey `json:"founders"`
	Chain    []ModSet            `json:"chain"`
}

// Registry holds the verified moderator sets of known communities.
type Registry struct {
	sets map[string]map[uint64]*ModSet
}

// LoadRegistry reads every *.json CommunityFile in fsys and verifies it. Any bad file
// fails the whole load: an unverifiable community must not silently disappear.
func LoadRegistry(fsys fs.FS) (*Registry, error) {
	names, err := fs.Glob(fsys, "*.json")
	if err != nil {
		return nil, err
	}
	r := &Registry{sets: map[string]map[uint64]*ModSet{}}
	for _, name := range names {
		b, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		var cf CommunityFile
		if err := json.Unmarshal(b, &cf); err != nil {
			return nil, fmt.Errorf("%s: %w", path.Base(name), err)
		}
		if err := r.Add(cf); err != nil {
			return nil, fmt.Errorf("%s: %w", path.Base(name), err)
		}
	}
	return r, nil
}

// Add verifies and registers one community.
func (r *Registry) Add(cf CommunityFile) error {
	if r.sets == nil {
		r.sets = map[string]map[uint64]*ModSet{}
	}
	head, err := VerifyModSetChain(cf.Founders, cf.Chain)
	if err != nil {
		return err
	}
	if _, dup := r.sets[head.Community]; dup {
		return fmt.Errorf("community %q defined twice", head.Community)
	}
	r.sets[head.Community] = map[uint64]*ModSet{head.Epoch: head}
	return nil
}

// Get returns the verified set for community at epoch.
func (r *Registry) Get(community string, epoch uint64) (*ModSet, bool) {
	s, ok := r.sets[community][epoch]
	return s, ok
}

// Communities lists known communities, sorted.
func (r *Registry) Communities() []string {
	out := make([]string, 0, len(r.sets))
	for c := range r.sets {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// Head returns the community's latest verified moderator set (the one authors collect
// votes from).
func (r *Registry) Head(community string) (*ModSet, bool) {
	var head *ModSet
	for _, s := range r.sets[community] {
		if head == nil || s.Epoch > head.Epoch {
			head = s
		}
	}
	return head, head != nil
}
