package engine

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/exp/maps"
	"log/slog"
	"mosaic/mosaicdb"
	"mosaic/types"
	"testing"
)

func TestSplitFileToChunks(t *testing.T) {
	// shard id : size in bytes
	estimateShardsSize := []uint64{11, 10, 10, 11}
	checkSplitFileToChunks(t, 7, buildShardState(), estimateShardsSize)

	estimateShardsSize = []uint64{10, 10, 10, 10}

	checkSplitFileToChunks(t, 5, buildShardState(), estimateShardsSize)

	estimateShardsSize = []uint64{10, 10, 10, 9}
	checkSplitFileToChunks(t, 4, buildShardState(), estimateShardsSize)

	estimateShardsSize = []uint64{11, 11, 12, 11}
	checkSplitFileToChunks(t, 10, buildShardState(), estimateShardsSize)

	estimateShardsSize = []uint64{11, 11, 11, 11}
	checkSplitFileToChunks(t, 7, buildShardState3(), estimateShardsSize)

	estimateShardsSize = []uint64{11, 11, 11, 12}
	checkSplitFileToChunks(t, 8, buildShardState3(), estimateShardsSize)
}

func buildShardState() map[types.ShardId]uint64 {
	shards := map[types.ShardId]uint64{
		1: 10,
		2: 10,
		3: 10,
		4: 5,
	}
	return shards
}

func buildShardState3() map[types.ShardId]uint64 {
	shards := map[types.ShardId]uint64{
		1: 10,
		2: 8,
		3: 8,
		4: 11,
	}
	return shards
}

func checkSplitFileToChunks(t *testing.T, fileSize uint64, shardsSize map[types.ShardId]uint64, estimateShardsSize []uint64) {
	slog.Info("File", "size", fileSize)
	slog.Info("Shards", "state", shardsSize)

	result := splitFileToChunks(fileSize, shardsSize)
	sum := uint64(0)
	for _, size := range maps.Values(result) {
		sum += size
	}
	if fileSize != sum {
		t.Fatal("Invalid file size")
	}

	for id, size := range result {
		shardsSize[id] += size
	}
	slog.Info("Shards", "state", shardsSize)

	for _, val := range estimateShardsSize {
		for k, v := range shardsSize {
			if v == val {
				delete(shardsSize, k)
				break
			}
		}
	}
	if len(shardsSize) > 0 {
		t.Fatal("Invalid shards sizes")
	}
}

func TestFileChunkToChunkInfo(t *testing.T) {
	chunksRaw := map[types.ShardId]uint64{1: 2, 2: 3, 3: 2, 4: 0, 5: 3}
	chunks := fileChunkToChunkInfo(chunksRaw)
	fileData, _ := hex.DecodeString("2e84bf84632dd0adedb4")
	slog.Info("file", "data", hex.EncodeToString(fileData))
	printMap(chunks)
	for _, chunkInfo := range chunks {
		val := hex.EncodeToString(splitFile(fileData, chunkInfo.Offset, chunkInfo.Fsize))
		/* TODO
		switch chunkInfo.ShardId {
		case 1:
			if "2e84" != val {
				t.Fatal("Err chunk size", "want", "2e84", "got", val)
			}
		case 2:
			if "bf8463" != val {
				t.Fatal("Err chunk size", "want", "bf8463", "got", val)
			}
		case 3:
			if "2dd0" != val {
				t.Fatal("Err chunk size", "want", "2dd0", "got", val)
			}
		case 5:
			if "adedb4" != val {
				t.Fatal("Err chunk size", "want", "adedb4", "got", val)
			}
		}
		*/
		slog.Info("data", "sId", chunkInfo.ShardId, "o", chunkInfo.Offset, "s", chunkInfo.Fsize, "data", val)
	}
}

func printMap(sl []*mosaicdb.ChunkInfo) {
	for _, val := range sl {
		slog.Info("Value", "obj", fmt.Sprintf("%+v", val))
	}
}
