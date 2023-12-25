package engine

import (
	"github.com/gofiber/fiber/v2/log"
	"log/slog"
)

func (cmd *FileUploadErrCmd) Prepare(engine *Engine) error {
	cmd.engine = engine
	return nil
}

// Execute we use simple logic to resolve upload err:
// remove file chunks from shards, decrease shards bytes size usage
// without file upload again, compensations etc.
func (cmd *FileUploadErrCmd) Execute() error {
	fileMeta := cmd.FileMetadata
	err := cmd.Err
	slog.Info("Try resolve upload err", "fId", fileMeta.Id.String(), "err", err.Error())
	// decrease used space by file
	if err := cmd.engine.db.UpdateShardsSize(fileMeta.UsedSpace(), false); err != nil {
		return err
	}
	// async free resources on shards
	go func() {
		for _, chunk := range fileMeta.Chunks {
			slog.Info("Start remove chunk", "sId", chunk.ShardId, "id", chunk.Id.String())
			if err := cmd.engine.storage.Remove(chunk.ShardId, chunk.Id); err != nil {
				slog.Error("Failed to remove chunk from shard", "sId", chunk.ShardId, "id", chunk.Id.String())
			}
			slog.Info("End remove chunk", "sId", chunk.ShardId, "id", chunk.Id.String())
		}

		if err := cmd.engine.removeTempFile(fileMeta.Id); err != nil {
			log.Errorf("Failed to remove temp file with id:%s err:%s", fileMeta.Id.String(), err.Error())
		}
	}()

	return nil
}
