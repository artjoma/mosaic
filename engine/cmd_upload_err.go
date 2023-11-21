package engine

import (
	"log/slog"
	"mosaic/mosaicdb"
)

// PrepareUploadErrCmd register upload err
func (engine *Engine) PrepareUploadErrCmd(fileMetadata *mosaicdb.FileMetadata, err error) {
	cmd := &FileUploadErrCmd{
		Err:          err,
		FileMetadata: fileMetadata,
	}

	engine.commandChan <- cmd
}

// FileUploadErrCmd we use simple logic to resolve upload err:
// remove file chunks from shards, decrease shards bytes size usage
// without file upload again, compensations etc.
func (engine *Engine) fileUploadErrCmd(cmd *FileUploadErrCmd) {
	fileMeta := cmd.FileMetadata
	err := cmd.Err
	slog.Info("Try resolve upload err", "fId", fileMeta.Id.String(), "err", err.Error())
	// decrease used space by file
	engine.db.UpdateShardsSize(fileMeta.UsedSpace(), false)
	// async free resources on shards
	go func() {
		for _, chunk := range fileMeta.Chunks {
			slog.Info("Start remove chunk", "sId", chunk.ShardId, "id", chunk.Id.String())
			if err := engine.storage.Remove(chunk.ShardId, chunk.Id); err != nil {
				slog.Error("Failed to remove chunk from shard", "sId", chunk.ShardId, "id", chunk.Id.String())
			}
			slog.Info("End remove chunk", "sId", chunk.ShardId, "id", chunk.Id.String())
		}

		engine.removeTempFile(fileMeta.Id)
	}()
}
