package dbnode

import (
	"sync"

	"github.com/libr-forum/Libr/core/libr/proto"
)

// MemStore is an in-memory Store for tests.
type MemStore struct {
	mu    sync.Mutex
	certs map[proto.Hash]proto.MsgCert
	index map[string][]proto.Hash
}

func NewMemStore() *MemStore {
	return &MemStore{certs: map[proto.Hash]proto.MsgCert{}, index: map[string][]proto.Hash{}}
}

func (m *MemStore) Put(h proto.Hash, c proto.MsgCert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.certs[h] = c
	return nil
}

func (m *MemStore) Get(h proto.Hash) (*proto.MsgCert, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.certs[h]; ok {
		return &c, nil
	}
	return nil, nil
}

func (m *MemStore) AddIndex(community string, bucket int64, h proto.Hash) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := indexKey(community, bucket)
	for _, x := range m.index[k] {
		if x == h {
			return nil
		}
	}
	m.index[k] = append(m.index[k], h)
	return nil
}

func (m *MemStore) Index(community string, bucket int64, limit int) ([]proto.Hash, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	hs := m.index[indexKey(community, bucket)]
	if len(hs) > limit {
		hs = hs[:limit]
	}
	return append([]proto.Hash(nil), hs...), nil
}

func indexKey(community string, bucket int64) string {
	return string(proto.Encode("", []byte(community), proto.I64(bucket)))
}
