package storage

import (
	"bytes"
	"fmt"
	"mosaic/mosaicdb"
	"mosaic/types"
	"mosaic/utils"
	"testing"
)

func TestStorage(t *testing.T) {
	shards := mosaicdb.NewShards()
	storage := NewStorage()

	// make all connections to peers
	var shardId types.ShardId
	for port := utils.TestPortRangeFrom; port <= utils.TestPortRangeTo; port++ {
		host := fmt.Sprintf("0.0.0.0:%d", port)
		shard := &mosaicdb.Shard{
			Host: host,
			Id:   mosaicdb.BuildShardId(host),
		}
		shards.AddShard(shard)
		storage.AddPeer(shard)
		shardId = shard.Id
	}
	data := utils.RandomBytes(1024 * 1024 * 2)
	id := utils.HashFrom(data)
	// test put
	if err := storage.Put(shardId, id, data); err != nil {
		t.Error(err)
	}
	// test if present by key
	present, err := storage.Has(shardId, id)
	if err != nil {
		t.Fatal(err)
	}
	if !present {
		t.Fatal("File id not found", id)
	}
	// test get
	_data, err := storage.Get(shardId, id)
	_id := utils.HashFrom(_data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(id, _id) {
		t.Fatal("Invalid hashes")
	}
}
