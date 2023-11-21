package engine

import (
	"golang.org/x/exp/maps"
	"mosaic/types"
	"slices"
	"sort"
)

// splitFileParts split file to multiple chunk according to shards state
// return shardId : file chunk size
func splitFileToChunks(fileSize uint64, shardsSize map[types.ShardId]uint64) map[types.ShardId]uint64 {
	shardCount := uint64(len(shardsSize))
	splitFileParts := make(map[types.ShardId]uint64, shardCount)
	// split remainder over all peers, related to peers Min(sizes) and Max(sizes)
	values := maps.Values(shardsSize)
	maxSize := slices.Max(values)
	//Sort ASC by shard size
	sizes := mapToSortedArray(shardsSize, 1)
	// totalDiff how many bytes we should add to all shards till shard with max size
	totalDiff := uint64(0)
	for _, size := range sizes {
		difShardVsMaxSize := maxSize - size[1]
		// how many bytes we should add to shard(N) till shard with max size
		if fileSize < totalDiff+difShardVsMaxSize {
			count := fileSize - totalDiff
			totalDiff += count
			size[2] = count
			break
		} else {
			totalDiff += difShardVsMaxSize
			size[2] = difShardVsMaxSize
		}
	}

	fileSizeNormalized := fileSize - totalDiff
	remainder := fileSizeNormalized % shardCount
	if remainder != 0 {
		for _, size := range sizes {
			if remainder > 0 {
				size[2] += 1
				remainder -= 1
			}
		}
	}
	fileSizeNormalized -= remainder
	splitSize := fileSizeNormalized / shardCount
	for _, size := range sizes {
		peerId := types.ShardId(size[0])
		diff := size[2]
		splitFileParts[peerId] = splitSize + diff
	}

	return splitFileParts
}

// MapToSortedArray Sort ASC, return [{shardId, size, diff is 0}]
func mapToSortedArray(data map[types.ShardId]uint64, sortCol int) [][]uint64 {
	var values [][]uint64
	for k, v := range data {
		values = append(values, []uint64{uint64(k), v, 0})
	}
	sort.Slice(values[:], func(i, j int) bool {
		return values[i][sortCol] < values[j][sortCol]
	})
	return values
}
