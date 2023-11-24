package engine

import (
	"bytes"
	"errors"
	"github.com/cockroachdb/pebble"
	"log/slog"
	"mosaic/mosaicdb"
	"mosaic/storage"
	"mosaic/types"
	"os"
	"path/filepath"
	"time"
)

type Engine struct {
	tempFolder     string
	db             *mosaicdb.Database
	storage        *storage.Storage
	commandChan    chan ProcessorCommand
	fileUploader   *FileUploader
	fileDownloader *FileDownloader
}

func NewEngine(tempFolder string, db *mosaicdb.Database, dataStorage *storage.Storage,
	uploadWorkersCount int) (*Engine, error) {
	// create folder if not exists
	err := os.MkdirAll(tempFolder, 0774)
	slog.Info("Create dir", "path", tempFolder, "err", err)
	engine := &Engine{
		tempFolder:     tempFolder,
		commandChan:    make(chan ProcessorCommand, 16),
		db:             db,
		storage:        dataStorage,
		fileDownloader: NewFileDownloader(dataStorage),
	}
	engine.fileUploader = NewFileUploader(engine, uploadWorkersCount)
	// prepare peers clients
	shards, err := engine.db.GetShards()
	if errors.Is(err, pebble.ErrNotFound) {
		slog.Info("Shards not found")
		// save empty lists
		db.SaveShards(mosaicdb.NewShards())
	} else {
		for _, shard := range shards.Shards {
			dataStorage.AddPeer(shard)
		}
	}
	go engine.startCommandListener()
	return engine, nil
}
func (engine *Engine) ExecuteCmdAsync(cmd ProcessorCommand) {
	engine.commandChan <- cmd
}

// Process all income commands in single thread.
// Try process commands a faster as possible.
// Try to prepare command/finalize command before/after putting command to queue
// Thinking twice before add new logic
func (engine *Engine) startCommandListener() {
	for cmd := range engine.commandChan {
		now := time.Now()
		if err := cmd.Execute(); err != nil {
			slog.Error("Failed to execute", "cmd", cmd.getCmdId(), "err", err.Error())
		}

		slog.Info("Processor", "cmd", cmd.getCmdId(), "took(ms)", time.Since(now).Milliseconds())
	}
}

// shardsSize return
func (engine *Engine) shardsSize() (map[types.ShardId]uint64, error) {
	shardsSize := make(map[types.ShardId]uint64)
	for _, id := range engine.storage.GetPeersId() {
		size, err := engine.db.GetShardSize(id)
		if err != nil {
			slog.Error("Failed to call engine.db.GetShardSize(id)", "id", id, "err", err)
			return nil, err
		}
		shardsSize[id] = size
	}
	return shardsSize, nil
}

func (engine *Engine) removeTempFile(fileId types.H256) error {
	slog.Info("Start remove from temp", "file", fileId.String())
	if err := os.Remove(filepath.Join(engine.tempFolder, fileId.String())); err != nil {
		return err
	}
	slog.Info("End remove from temp", "file", fileId.String())
	return nil
}

func (engine *Engine) DownloadFile(fileId types.H256) (*mosaicdb.FileMetadata, *bytes.Buffer, error) {
	fMeta, err := engine.db.GetFileMetadataById(fileId)
	if err != nil {
		return nil, nil, err
	}
	if fMeta.Status != mosaicdb.FileStatusReady {
		return fMeta, nil, errors.New("file not ready")
	}

	now := time.Now()
	slog.Info("Try download", "file", fileId.String())
	data, err := engine.fileDownloader.downloadFile(fMeta)
	if err != nil {
		slog.Error("Failed to download", "fileId", fileId.String(), "err", err.Error())
		return fMeta, nil, err
	}
	slog.Info("File ready", "id", fileId.String(), "took(ms)", time.Since(now).Milliseconds())

	return fMeta, data, err
}

func (engine *Engine) GetFileMetadata(fileId types.H256) (*mosaicdb.FileMetadata, error) {
	return engine.db.GetFileMetadataById(fileId)
}

func (engine *Engine) ClusterState() (map[types.ShardId]uint64, error) {
	ids := engine.storage.ClusterShardIds()
	return engine.db.GetShardsSize(ids)
}
