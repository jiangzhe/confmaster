package template

type VersionSnapshotStore interface {
	SnapshotStore

	// given a specified version, return the corresponding snapshot
	VersionedSnapshot(version string) (VersionedSnapshot, error)
}

type VersionedSnapshot interface {
	Snapshot

	// return the version of this kvs snapshot
	Version() string
}