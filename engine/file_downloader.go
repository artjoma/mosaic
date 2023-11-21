package engine

import (
	"bytes"
	"errors"
	"log/slog"
	"mosaic/mosaicdb"
	"mosaic/storage"
	"mosaic/utils"
	"sync"
)

type FileDownloader struct {
	storage *storage.Storage
}

func NewFileDownloader(storage *storage.Storage) *FileDownloader {
	return &FileDownloader{
		storage: storage,
	}
}

func (downloader *FileDownloader) downloadFile(fileMeta *mosaicdb.FileMetadata) (*bytes.Buffer, error) {
	workers := sync.WaitGroup{}
	workers.Add(len(fileMeta.Chunks))
	var downloadErr error
	parts := sync.Map{}

	for workerId, chunk := range fileMeta.Chunks {
		go func(workerId int, chunk *mosaicdb.ChunkInfo) {
			defer workers.Done()
			slog.Info("Download chunk", "worker", workerId, "shardId", chunk.ShardId, "id", chunk.Id.String(), "size", chunk.Fsize)
			chunkData, err := downloader.storage.Get(chunk.ShardId, chunk.Id)
			if !bytes.Equal(chunk.Id, utils.HashFrom(chunkData)) {
				slog.Error("Invalid chunk hash", "id", chunk.Id.String(), "got", utils.HashFrom(chunkData).String())
				err = errors.New("invalid chunk hash")
			}

			if err == nil {
				parts.Store(chunk.ShardId, chunkData)
			} else {
				downloadErr = err
			}
		}(workerId, chunk)
	}
	workers.Wait()

	if downloadErr != nil {
		return nil, downloadErr
	}

	buff := &bytes.Buffer{}
	// merge chunks to single file
	for _, chunk := range fileMeta.Chunks {
		val, _ := parts.Load(chunk.ShardId)
		buff.Write(val.([]byte))
	}
	data := buff.Bytes()
	if fileMeta.FileSize != uint64(len(data)) {
		return nil, errors.New("invalid file length")
	}

	return buff, nil
}
