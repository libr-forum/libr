// Package sqlstore is the SQLite-backed dbnode.Store used by the DB node binary.
package sqlstore

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/libr-forum/Libr/core/libr/proto"
)

// Store is a dbnode.Store on one SQLite file.
type Store struct{ db *sql.DB }

const schema = `
CREATE TABLE IF NOT EXISTS certs (
	hash      BLOB PRIMARY KEY,
	msg_id    BLOB NOT NULL,
	community TEXT NOT NULL,
	ts        INTEGER NOT NULL,
	body      BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS certs_msg_id ON certs(msg_id);
CREATE TABLE IF NOT EXISTS feed_index (
	community TEXT NOT NULL,
	bucket    INTEGER NOT NULL,
	hash      BLOB NOT NULL,
	msg_id    BLOB NOT NULL,
	PRIMARY KEY (community, bucket, hash)
);
CREATE INDEX IF NOT EXISTS feed_index_msg_id ON feed_index(msg_id);
CREATE TABLE IF NOT EXISTS tombstones (
	msg_id    BLOB PRIMARY KEY,
	community TEXT NOT NULL,
	bucket    INTEGER NOT NULL,
	body      BLOB NOT NULL
);`

// Open opens (creating if needed) the store at path.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlstore schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Put(h proto.Hash, c proto.MsgCert) error {
	body, err := json.Marshal(c)
	if err != nil {
		return err
	}
	id := c.Msg.ID()
	_, err = s.db.Exec(`INSERT OR IGNORE INTO certs(hash, msg_id, community, ts, body) VALUES (?, ?, ?, ?, ?)`,
		h[:], id[:], c.Msg.Community, c.Msg.Ts, body)
	return err
}

func (s *Store) Get(h proto.Hash) (*proto.MsgCert, error) {
	var body []byte
	switch err := s.db.QueryRow(`SELECT body FROM certs WHERE hash = ?`, h[:]).Scan(&body); err {
	case nil:
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
	var c proto.MsgCert
	if err := json.Unmarshal(body, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) AddIndex(community string, bucket int64, h, msgID proto.Hash) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO feed_index(community, bucket, hash, msg_id) VALUES (?, ?, ?, ?)`,
		community, bucket, h[:], msgID[:])
	return err
}

// Remove records the tombstone and deletes the message's certificates and index
// entries in one transaction.
func (s *Store) Remove(t proto.Tombstone, community string, bucket int64) error {
	body, err := json.Marshal(t)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit
	for _, q := range []struct {
		sql  string
		args []any
	}{
		{`INSERT OR IGNORE INTO tombstones(msg_id, community, bucket, body) VALUES (?, ?, ?, ?)`, []any{t.MsgID[:], community, bucket, body}},
		{`DELETE FROM certs WHERE msg_id = ?`, []any{t.MsgID[:]}},
		{`DELETE FROM feed_index WHERE msg_id = ?`, []any{t.MsgID[:]}},
	} {
		if _, err := tx.Exec(q.sql, q.args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) Removed(msgID proto.Hash) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM tombstones WHERE msg_id = ?`, msgID[:]).Scan(&n)
	return n > 0, err
}

func (s *Store) Tombstones(community string, bucket int64, limit int) ([]proto.Tombstone, error) {
	rows, err := s.db.Query(`SELECT body FROM tombstones WHERE community = ? AND bucket = ? LIMIT ?`, community, bucket, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []proto.Tombstone
	for rows.Next() {
		var body []byte
		if err := rows.Scan(&body); err != nil {
			return nil, err
		}
		var t proto.Tombstone
		if err := json.Unmarshal(body, &t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) Index(community string, bucket int64, limit int) ([]proto.Hash, error) {
	rows, err := s.db.Query(`SELECT hash FROM feed_index WHERE community = ? AND bucket = ? LIMIT ?`, community, bucket, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []proto.Hash
	for rows.Next() {
		var b []byte
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}
		var h proto.Hash
		if copy(h[:], b) != len(h) {
			return nil, fmt.Errorf("sqlstore: bad hash length %d", len(b))
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
