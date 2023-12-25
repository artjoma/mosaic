package mosaicdb

import (
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"log/slog"
	"mosaic/types"
	"mosaic/utils"
	"os"
)

// static keys
var keyShards = []byte("shards")
var keyPrefixShardSize = []byte("s")
var keyPrefixMetaFile = []byte("m")

type Database struct {
	db *pebble.DB
}

// NewDatabase open local DB
// Use empty path for memory db!
func NewDatabase(dbPath string) *Database {
	var db *pebble.DB = nil
	var err error = nil
	// TODO tune settings according LSM specific
	if dbPath == "" {
		slog.Info("Initialize in mem DB")
		db, err = pebble.Open("", &pebble.Options{
			FS: vfs.NewMem(),
		})
	} else {
		slog.Info("Initialize db", "path", dbPath)
		os.MkdirAll(dbPath, 0774)
		db, err = pebble.Open(dbPath, &pebble.Options{})
	}
	if err != nil {
		panic(err)
	}
	return &Database{
		db: db,
	}
}

// SaveFileMetadata , add file metadata + increase shards used space
func (db *Database) SaveFileMetadata(metadata *FileMetadata) error {
	if data, err := utils.ToBytes(metadata); err == nil {
		key := db.BuildKey(keyPrefixMetaFile, metadata.Id)
		return db.db.Set(key, data, pebble.Sync)
	} else {
		return err
	}
}

func (db *Database) GetFileMetadataById(id types.H256) (*FileMetadata, error) {
	key := db.BuildKey(keyPrefixMetaFile, id)
	data, closer, err := db.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	fMeta := &FileMetadata{}
	return fMeta, utils.FromBytes(data, fMeta)
}

func (db *Database) AllMetadata() ([]types.H256, error) {
	iter, err := db.db.NewIter(prefixIterOptions([]byte("m")))
	fMetaIds := make([]types.H256, 0, 1_000)
	if err != nil {
		return fMetaIds, nil
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()[1:]
		cp := make([]byte, len(key))
		copy(cp, key)
		fMetaIds = append(fMetaIds, cp)
	}

	return fMetaIds, nil
}

func (db *Database) SaveShards(shards *Shards) error {
	if data, err := utils.ToBytes(shards); err == nil {
		return db.db.Set(keyShards, data, pebble.Sync)
	} else {
		return err
	}
}

func (db *Database) GetShards() (*Shards, error) {
	data, closer, err := db.db.Get(keyShards)
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	shards := &Shards{}
	return shards, utils.FromBytes(data, shards)
}

func (db *Database) GetShardsSize(ids []types.ShardId) (map[types.ShardId]uint64, error) {
	spaceUsage := make(map[types.ShardId]uint64)
	for _, id := range ids {
		size, err := db.GetShardSize(id)
		if err != nil {
			return nil, err
		}
		spaceUsage[id] = size
	}
	return spaceUsage, nil
}

// GetShardSize Return used space by shard
func (db *Database) GetShardSize(id types.ShardId) (uint64, error) {
	key := db.BuildKey(keyPrefixShardSize, utils.U32tob(uint32(id)))
	data, closer, err := db.db.Get(key)
	if err != nil {
		return 0, err
	}
	defer closer.Close()
	size := utils.Btou64(data)
	return size, nil
}

func (db *Database) UpdateShardsSize(diffs map[types.ShardId]uint64, increaseSize bool) error {
	for id, diff := range diffs {
		if err := db.UpdateShardSize(id, diff, increaseSize); err != nil {
			return err
		}
	}
	return nil
}

func (db *Database) UpdateShardSize(id types.ShardId, diff uint64, increaseSize bool) error {
	size, err := db.GetShardSize(id)
	if err != nil {
		return err
	}
	if increaseSize {
		size += diff
	} else {
		size -= diff
	}
	return db.SaveShardSize(id, size)
}

func (db *Database) ResetShardSize(shardIds []types.ShardId) error {
	for _, shardId := range shardIds {
		if err := db.SaveShardSize(shardId, 0); err != nil {
			return err
		}
	}
	return nil
}

func (db *Database) SaveShardSize(id types.ShardId, size uint64) error {
	key := db.BuildKey(keyPrefixShardSize, utils.U32tob(uint32(id)))
	return db.db.Set(key, utils.U64tob(size), pebble.Sync)
}

func (db *Database) BuildKey(prefix []byte, key []byte) []byte {
	return append(prefix, key...)
}

var keyUpperBound = func(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // no upper-bound
}

var prefixIterOptions = func(prefix []byte) *pebble.IterOptions {
	return &pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: keyUpperBound(prefix),
	}
}
