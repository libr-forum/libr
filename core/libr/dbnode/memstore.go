package dbnode

import (
	"sync"

	"github.com/libr-forum/Libr/core/libr/proto"
)

// MemStore is an in-memory Store for tests.
type MemStore struct {
	mu    sync.Mutex
	certs map[proto.Hash]proto.MsgCert
	index map[string][]entry
	tombs map[proto.Hash]proto.Tombstone
	byBkt map[string][]proto.Hash // tombstone msg IDs per bucket
}

type entry struct{ hash, msgID proto.Hash }

func NewMemStore() *MemStore {
	return &MemStore{
		certs: map[proto.Hash]proto.MsgCert{},
		index: map[string][]entry{},
		tombs: map[proto.Hash]proto.Tombstone{},
		byBkt: map[string][]proto.Hash{},
	}
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

func (m *MemStore) AddIndex(community string, bucket int64, h, msgID proto.Hash) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := indexKey(community, bucket)
	for _, x := range m.index[k] {
		if x.hash == h {
			return nil
		}
	}
	m.index[k] = append(m.index[k], entry{h, msgID})
	return nil
}

func (m *MemStore) Index(community string, bucket int64, limit int) ([]proto.Hash, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []proto.Hash
	for _, e := range m.index[indexKey(community, bucket)] {
		if len(out) == limit {
			break
		}
		out = append(out, e.hash)
	}
	return out, nil
}

func (m *MemStore) Remove(t proto.Tombstone, community string, bucket int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tombs[t.MsgID]; !ok {
		m.tombs[t.MsgID] = t
		k := indexKey(community, bucket)
		m.byBkt[k] = append(m.byBkt[k], t.MsgID)
	}
	for h, c := range m.certs {
		if c.Msg.ID() == t.MsgID {
			delete(m.certs, h)
		}
	}
	for k, es := range m.index {
		kept := es[:0]
		for _, e := range es {
			if e.msgID != t.MsgID {
				kept = append(kept, e)
			}
		}
		m.index[k] = kept
	}
	return nil
}

func (m *MemStore) Removed(msgID proto.Hash) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.tombs[msgID]
	return ok, nil
}

func (m *MemStore) Tombstones(community string, bucket int64, limit int) ([]proto.Tombstone, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []proto.Tombstone
	for _, id := range m.byBkt[indexKey(community, bucket)] {
		if len(out) == limit {
			break
		}
		out = append(out, m.tombs[id])
	}
	return out, nil
}

func indexKey(community string, bucket int64) string {
	return string(proto.Encode("", []byte(community), proto.I64(bucket)))
}
