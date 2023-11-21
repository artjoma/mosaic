package engine

import (
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"log/slog"
	"mosaic/mosaicdb"
	"mosaic/types"
	"sync"
	"time"
)

func (engine *Engine) addShardsCmd(cmd *AddShardCmd) error {
	slog.Info("Start create new shard", "host", cmd.Host)
	shards, err := engine.db.GetShards()
	if err != nil {
		return err
	}
	// create shard
	shard := &mosaicdb.Shard{
		Id:   cmd.Id,
		Host: cmd.Host,
	}
	// add shard to shards
	shards.AddShard(shard)
	// create client
	engine.storage.AddPeer(shard)
	// save list of peers
	if err := engine.db.SaveShards(shards); err != nil {
		return err
	}
	// create record with shard size = 0
	if err := engine.db.ResetShardSize(maps.Keys(shards.Shards)); err != nil {
		return err
	}
	slog.Info("Start analyze files count...")
	// get all files ids
	ids, err := engine.db.AllMetadata()
	if err != nil {
		return err
	}
	if len(ids) > 0 {
		slog.Info("Start shards rebalancing", "len", len(ids))
		// start download files one by one and upload again according to new cluster state
		go func() {
			now := time.Now()
			for i, id := range ids {
				if i%1000 == 0 {
					slog.Info("Rebalanced file", "n", i, "from", len(ids))
				}
				// download
				fMeta, fData, err := engine.DownloadFile(id)
				if err != nil {
					slog.Error("Failed downloadFile()", "id", id.String(), "err", err.Error())
				}
				// async remove file parts from shards
				workers := sync.WaitGroup{}
				workers.Add(len(fMeta.Chunks))
				for _, chunk := range fMeta.Chunks {
					go func() {
						defer workers.Done()
						engine.storage.Remove(chunk.ShardId, chunk.Id)
					}()
				}
				// wait until done
				workers.Wait()

				engine.commandChan <- &PutFileCmd{
					FileMetadata: fMeta,
					FileData:     fData.Bytes(),
				}
			}
			slog.Info("End shards rebalancing", "took", time.Since(now))
		}()
	}

	slog.Info("End create new shard", "host", cmd.Host, "id", shard.Id)
	return nil
}

func (engine *Engine) PrepareAddShardsCmd(host string) (types.ShardId, error) {
	shardId := mosaicdb.BuildShardId(host)
	found := engine.storage.HasShard(shardId)
	if found {
		return 0, errors.New(fmt.Sprintf("shard with id:%d already reistered", shardId))
	}
	engine.commandChan <- &AddShardCmd{
		Host: host,
		Id:   shardId,
	}

	return shardId, nil
}
