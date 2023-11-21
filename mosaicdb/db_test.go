package mosaicdb

import (
	"mosaic/types"
	"mosaic/utils"
	"testing"
)

func TestSaveShards(t *testing.T) {
	shards := NewShards()

	shard1 := shards.AddShard(createShard("0.0.0.0:9995")).Id
	shard2 := shards.AddShard(createShard("0.0.0.0:9997")).Id
	shard3 := shards.AddShard(createShard("0.0.0.0:9999")).Id
	updatedAt := shards.UpdatedAt
	db := NewDatabase("")
	if err := db.SaveShards(shards); err != nil {
		t.Fatal(err)
	}

	_shards, err := db.GetShards()
	if err != nil {
		t.Fatal(err)
	}
	if _shards.UpdatedAt != updatedAt {
		t.Fatal("Invalid updatedAt field")
	}
	if _shards.GetShardById(shard1).Host != "0.0.0.0:9995" {
		t.Fatal("Invalid shard1 host")
	}
	if _shards.GetShardById(shard2).Host != "0.0.0.0:9997" {
		t.Fatal("Invalid shard1 host")
	}
	if _shards.GetShardById(shard3).Host != "0.0.0.0:9999" {
		t.Fatal("Invalid shard1 host")
	}
}

func createShard(host string) *Shard {
	return &Shard{
		Host: "0.0.0.0:9995",
	}
}

func TestReadByPrefix(t *testing.T) {
	db := NewDatabase("")
	count := 10
	for i := 0; i < count; i++ {
		id := utils.HashFrom(utils.RandomBytes(128))
		fMeta := &FileMetadata{
			Id: id,
		}
		if err := db.SaveFileMetadata(fMeta); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 10; i++ {
		db.SaveShardSize(types.ShardId(i), 100)
	}
	ids, err := db.AllMetadata()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != count {
		t.Fatal("Invalid fMeta count", "actual", count, "got", len(ids))
	}
}
