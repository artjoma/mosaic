package engine

import (
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	"golang.org/x/exp/maps"
	"log/slog"
	"mosaic/mosaicdb"
	"sync"
	"time"
)

func (cmd *AddShardCmd) Execute() error {
	slog.Info("Start create new shard", "host", cmd.Host)
	shards, err := cmd.engine.db.GetShards()
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
	cmd.engine.storage.AddPeer(shard)
	// save list of peers
	if err := cmd.engine.db.SaveShards(shards); err != nil {
		return err
	}
	// create record with shard size = 0
	if err := cmd.engine.db.ResetShardSize(maps.Keys(shards.Shards)); err != nil {
		return err
	}
	slog.Info("Start analyze files count...")
	// get all files ids
	ids, err := cmd.engine.db.AllMetadata()
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
				fMeta, fData, err := cmd.engine.DownloadFile(id)
				if err != nil {
					slog.Error("Failed downloadFile()", "id", id.String(), "err", err.Error())
				}
				// async remove file parts from shards
				workers := sync.WaitGroup{}
				workers.Add(len(fMeta.Chunks))
				for _, chunkModel := range fMeta.Chunks {
					go func(chunk *mosaicdb.ChunkInfo) {
						defer workers.Done()
						if err := cmd.engine.storage.Remove(chunk.ShardId, chunk.Id); err != nil {
							log.Errorf("Faield to remove shardId:%d chunkId:%s err:%s", chunk.ShardId, chunk.Id.String(), err.Error())
						}
					}(chunkModel)
				}
				// wait until done
				workers.Wait()

				cmd.engine.ExecuteCmdAsync(&PutFileCmd{
					engine:       cmd.engine,
					FileMetadata: fMeta,
					FileData:     fData.Bytes(),
				})
			}

			slog.Info("End shards rebalancing", "took", time.Since(now))
		}()
	}

	slog.Info("End create new shard", "host", cmd.Host, "id", shard.Id)
	return nil
}

func (cmd *AddShardCmd) Prepare(engine *Engine) error {
	cmd.engine = engine
	shardId := mosaicdb.BuildShardId(cmd.Host)
	found := cmd.engine.storage.HasShard(shardId)
	cmd.Id = shardId
	if found {
		return fmt.Errorf("shard with id:%d already registered", shardId)
	}

	return nil
}
