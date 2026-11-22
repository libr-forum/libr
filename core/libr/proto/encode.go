// Package proto defines LIBR v2's signed objects and the exact bytes each signature
// covers. DB nodes, desktop and mobile all verify with this package, so the encoding is
// identical everywhere by construction.
package proto

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// Domains separate signature contexts: a signature for one object type can never be
// replayed as another.
const (
	DomainMsg          = "libr/v2/msg"
	DomainModVote      = "libr/v2/modvote"
	DomainMsgCert      = "libr/v2/msgcert"
	DomainRepVote      = "libr/v2/repvote"
	DomainAuthorDelete = "libr/v2/authordelete"
	DomainReceipt      = "libr/v2/receipt"
	DomainModSet       = "libr/v2/modset"
)

// Encode returns the canonical bytes for a signed object: the domain and then each
// field, every one prefixed with its uvarint length. Length prefixes make the encoding
// injective, unlike the v1 "content+ts+status" concatenation.
func Encode(domain string, fields ...[]byte) []byte {
	n := len(domain) + binary.MaxVarintLen64
	for _, f := range fields {
		n += len(f) + binary.MaxVarintLen64
	}
	b := make([]byte, 0, n)
	b = binary.AppendUvarint(b, uint64(len(domain)))
	b = append(b, domain...)
	for _, f := range fields {
		b = binary.AppendUvarint(b, uint64(len(f)))
		b = append(b, f...)
	}
	return b
}

// U64 is a fixed-width big-endian field.
func U64(v uint64) []byte { return binary.BigEndian.AppendUint64(nil, v) }

// I64 is a fixed-width big-endian field (two's complement).
func I64(v int64) []byte { return U64(uint64(v)) }

// Bool is a one-byte field.
func Bool(v bool) []byte {
	if v {
		return []byte{1}
	}
	return []byte{0}
}

// Hash is a SHA-256 digest. It marshals as hex.
type Hash [32]byte

// Sum hashes b.
func Sum(b []byte) Hash { return sha256.Sum256(b) }

func (h Hash) String() string { return hex.EncodeToString(h[:]) }

// IsZero reports whether h is all zeros (e.g. a genesis PrevHash).
func (h Hash) IsZero() bool { return h == Hash{} }

func (h Hash) MarshalText() ([]byte, error) { return []byte(h.String()), nil }

func (h *Hash) UnmarshalText(b []byte) error {
	if len(b) != 2*len(h) {
		return fmt.Errorf("hash: want %d hex chars, got %d", 2*len(h), len(b))
	}
	_, err := hex.Decode(h[:], b)
	return err
}
