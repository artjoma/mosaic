package mosaicdb

import (
	"encoding/json"
	"github.com/twmb/murmur3"
	"mosaic/types"
	"time"
)

type FileMetadata struct {
	Id               types.H256   `msg:"id"`     // hashOf(original file), key of KV
	OriginalFileName string       `msg:"oFName"` // original file name
	OriginalFileSize int64        `msg:"oFSize"`
	FileSize         uint64       `msg:"size"`
	Status           FileStatus   `msg:"status"`
	Chunks           []*ChunkInfo `msg:"chunks"`
}

// UsedSpace return nil if chunks is nil
func (fMeta *FileMetadata) UsedSpace() map[types.ShardId]uint64 {
	sizes := make(map[types.ShardId]uint64)
	if fMeta.Chunks == nil {
		return nil
	}
	for _, chunk := range fMeta.Chunks {
		sizes[chunk.ShardId] = chunk.Fsize
	}

	return sizes
}

type FileStatus uint8

const FileStatusPending FileStatus = 1
const FileStatusUploading FileStatus = 2
const FileStatusReady FileStatus = 3
const FileStatusErr FileStatus = 4

func (status FileStatus) String() string {
	switch status {
	case FileStatusPending:
		return "pending"
	case FileStatusUploading:
		return "uploading"
	case FileStatusReady:
		return "ready"
	case FileStatusErr:
		return "error"
	default:
		return "?"
	}
}

func (status FileStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

type ChunkInfo struct {
	Id      types.H256    `msg:"id"`      // chunk hash, key of kv
	ShardId types.ShardId `msg:"shardId"` // server shard id
	Offset  uint64        `msg:"offset"`  // Offset at full file
	Fsize   uint64        `msg:"fSize"`   // chunk file size
}

// Shards store information about shards
type Shards struct {
	UpdatedAt int64                    `msg:"updAt"`  // Unix time
	Shards    map[types.ShardId]*Shard `msg:"shards"` // shards id:address, Maps must have string keys (MsgPack recommendation)
}

func NewShards() *Shards {
	return &Shards{
		UpdatedAt: time.Now().Unix(),
		Shards:    make(map[types.ShardId]*Shard),
	}
}

type Shard struct {
	Id   types.ShardId
	Host string
}

func (shards *Shards) AddShard(shard *Shard) *Shard {
	shards.Shards[shard.Id] = shard
	shards.UpdatedAt = time.Now().Unix()
	return shard
}

func BuildShardId(host string) types.ShardId {
	var h32 = murmur3.New32()
	h32.Write([]byte(host))
	return types.ShardId(h32.Sum32())
}

// GetShardById return nil if not found
func (shards *Shards) GetShardById(id types.ShardId) *Shard {
	if shard, ok := shards.Shards[id]; ok {
		return shard
	} else {
		return nil
	}
}
