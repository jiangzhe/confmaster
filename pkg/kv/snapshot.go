package kv

import "errors"

// The SnapshotStore is the general entry point to
// view and modify key values
type SnapshotStore interface {
	Snapshot() (Snapshot, error)
}

var (
	ErrSnapshotNotExists = errors.New("snapshot not exists")
	ErrSnapshotUnavailable = errors.New("snapshot fetch unavailable")
	ErrSnapshotCommitConflict = errors.New("snapshot commit conflict")
	ErrSnapshotCommitUnavailable = errors.New("snapshot commit unavailable")
)

// SCN is a value that specify the kv version
// it will be returned with a SnapshotStore
// and prevent concurrent modification
// for implementation, it could be Etag(http)
// or ResourceVersion(kubernetes), etc.
type SCN string

type Snapshot interface {
	// version of the kvs snapshot
	SCN() SCN

	// the store of kvs
	Config() Config

	// commit the change of kvs, error if conflict with version
	Commit() error
}

