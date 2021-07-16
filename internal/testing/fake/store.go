package fake

import (
	"go.dedis.ch/dela/core/store"
	"go.dedis.ch/dela/core/store/kv"
)

// InMemorySnapshot is a fake implementation of a store snapshot.
//
// - implements store.Snapshot
type InMemorySnapshot struct {
	store.Snapshot

	values    map[string][]byte
	ErrRead   error
	ErrWrite  error
	ErrDelete error
}

// NewSnapshot creates a new empty snapshot.
func NewSnapshot() *InMemorySnapshot {
	return &InMemorySnapshot{
		values: make(map[string][]byte),
	}
}

// NewBadSnapshot creates a new empty snapshot that will always return an error.
func NewBadSnapshot() *InMemorySnapshot {
	return &InMemorySnapshot{
		values:    make(map[string][]byte),
		ErrRead:   fakeErr,
		ErrWrite:  fakeErr,
		ErrDelete: fakeErr,
	}
}

// Get implements store.Snapshot.
func (snap *InMemorySnapshot) Get(key []byte) ([]byte, error) {
	return snap.values[string(key)], snap.ErrRead
}

// Set implements store.Snapshot.
func (snap *InMemorySnapshot) Set(key, value []byte) error {
	snap.values[string(key)] = value

	return snap.ErrWrite
}

// Delete implements store.Snapshot.
func (snap *InMemorySnapshot) Delete(key []byte) error {
	delete(snap.values, string(key))

	return snap.ErrDelete
}

// InMemoryDB is a fake implementation of a key/value storage.
//
// - implements kv.DB
type InMemoryDB struct {
	buckets map[string]*Bucket
	err     error
	errView error
}

// NewInMemoryDB returns a new empty database.
func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		buckets: make(map[string]*Bucket),
	}
}

// NewBadDB returns a new database that will return an error inside the
// transactions, and when closing the database.
func NewBadDB() *InMemoryDB {
	db := NewInMemoryDB()
	db.err = fakeErr

	return db
}

// NewBadViewDB returns a new database that fails to open view transactions.
func NewBadViewDB() *InMemoryDB {
	db := NewInMemoryDB()
	db.errView = fakeErr

	return db
}

// SetBucket allows to define a bucket in the database.
func (db *InMemoryDB) SetBucket(name []byte, b *Bucket) {
	db.buckets[string(name)] = b
}

// View implements kv.DB.
func (db *InMemoryDB) View(fn func(tx kv.ReadableTx) error) error {
	if db.errView != nil {
		return db.errView
	}

	return fn(dbTx{buckets: db.buckets, err: db.err})
}

// Update implements kv.DB.
func (db *InMemoryDB) Update(fn func(tx kv.WritableTx) error) error {
	return fn(dbTx{buckets: db.buckets, err: db.err})
}

// Close implements kv.DB.
func (db *InMemoryDB) Close() error {
	return db.err
}

// dbTx is a fake implementation of a database transaction.
//
// - implements kv.WritableTx
type dbTx struct {
	buckets map[string]*Bucket
	err     error
}

// GetBucket implements kv.ReadableTx.
func (tx dbTx) GetBucket(name []byte) kv.Bucket {
	bucket, found := tx.buckets[string(name)]
	if !found {
		return nil
	}

	return bucket
}

// GetBucketOrCreate implements kv.WritableTx.
func (tx dbTx) GetBucketOrCreate(name []byte) (kv.Bucket, error) {
	return tx.GetBucket(name), tx.err
}

// GetBucketOrCreate implements store.Transaction.
func (dbTx) OnCommit(fn func()) {}

// Bucket is a fake key/value storage bucket.
//
// - implements kv.Bucket
type Bucket struct {
	kv.Bucket

	values    map[string][]byte
	errSet    error
	errDelete error
}

// NewBucket returns a new empty bucket.
func NewBucket() *Bucket {
	return &Bucket{
		values: make(map[string][]byte),
	}
}

// NewBadWriteBucket returns a new empty bucket that fails to write.
func NewBadWriteBucket() *Bucket {
	b := NewBucket()
	b.errSet = fakeErr

	return b
}

// NewBadDeleteBucket returns a new empty bucket that fails to delete.
func NewBadDeleteBucket() *Bucket {
	b := NewBucket()
	b.errDelete = fakeErr

	return b
}

// Get implements kv.Bucket.
func (b *Bucket) Get(key []byte) []byte {
	return b.values[string(key)]
}

// Set implements kv.Bucket.
func (b *Bucket) Set(key, value []byte) error {
	b.values[string(key)] = value

	return b.errSet
}

// Delete implements kv.Bucket.
func (b *Bucket) Delete(key []byte) error {
	delete(b.values, string(key))

	return b.errDelete
}

// ForEach implements kv.Bucket.
func (b *Bucket) ForEach(fn func(key, value []byte) error) error {
	for key, value := range b.values {
		err := fn([]byte(key), value)
		if err != nil {
			return err
		}
	}

	return nil
}
