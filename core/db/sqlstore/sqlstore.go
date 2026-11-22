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
	PRIMARY KEY (community, bucket, hash)
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

func (s *Store) AddIndex(community string, bucket int64, h proto.Hash) error {
	_, err := s.db.Exec(`INSERT OR IGNORE INTO feed_index(community, bucket, hash) VALUES (?, ?, ?)`, community, bucket, h[:])
	return err
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
