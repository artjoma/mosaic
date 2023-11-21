package storage

import (
	"context"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/maps"
	"log/slog"
	"mosaic/mosaicdb"
	"mosaic/types"
	"strconv"
)

type Storage struct {
	peers map[types.ShardId]*Peer
}

type Peer struct {
	shard  *mosaicdb.Shard
	client *redis.Client
}

func NewStorage() *Storage {
	return &Storage{
		peers: make(map[types.ShardId]*Peer),
	}
}

func (storage *Storage) GetPeersId() []types.ShardId {
	return maps.Keys(storage.peers)
}

func (storage *Storage) AddPeer(shard *mosaicdb.Shard) {
	slog.Info("Start create peer connection", "id", shard.Id, "host", shard.Host)
	client := redis.NewClient(&redis.Options{
		Addr:         shard.Host,
		Password:     "", // no password set
		DB:           0,  // use default DB
		WriteTimeout: -2,
		ClientName:   strconv.FormatUint(uint64(shard.Id), 10),
	})

	peer := &Peer{
		shard:  shard,
		client: client,
	}

	storage.peers[shard.Id] = peer
	slog.Info("End create peer connection", "id", shard.Id, "host", shard.Host, "client", client.String())
}

func (storage *Storage) GetShardCount() int {
	return len(storage.peers)
}

func (storage *Storage) Get(shardId types.ShardId, key types.H256) ([]byte, error) {
	client := storage.peers[shardId].client
	data := client.Get(context.TODO(), string(key))
	if data.Err() != nil {
		return nil, data.Err()
	}
	return data.Bytes()
}

// Has O(N) where N is the number of keys to check.
func (storage *Storage) Has(shardId types.ShardId, key types.H256) (bool, error) {
	client := storage.peers[shardId].client
	result := client.Exists(context.TODO(), string(key))
	if result.Err() != nil {
		return false, result.Err()
	}
	if val, err := result.Uint64(); err == nil {
		// 1-found, 0-not
		return val == 1, nil
	} else {
		return false, err
	}
}

func (storage *Storage) HasShard(shardId types.ShardId) bool {
	_, ok := storage.peers[shardId]
	return ok
}

func (storage *Storage) Put(shardId types.ShardId, key types.H256, data []byte) error {
	client := storage.peers[shardId].client
	return client.Set(context.TODO(), string(key), data, 0).Err()
}

func (storage *Storage) Remove(shardId types.ShardId, key types.H256) error {
	client := storage.peers[shardId].client
	return client.Del(context.TODO(), string(key)).Err()
}

func (storage *Storage) ClusterShardIds() []types.ShardId {
	return maps.Keys(storage.peers)
}
