package engine

import (
	"log/slog"
	"mosaic/mosaicdb"
	"mosaic/utils"
	"sync"
	"time"
)

type FileUploader struct {
	engine               *Engine
	workers              []chan *PutFileCmd
	uploadFileMutex      sync.Mutex
	uploadedFilesCounter uint64
	uploadWorkersCount   int
}

func NewFileUploader(engine *Engine, uploadWorkersCount int) *FileUploader {
	uploader := &FileUploader{
		engine:               engine,
		uploadedFilesCounter: 0,
		workers:              make([]chan *PutFileCmd, uploadWorkersCount),
		uploadWorkersCount:   uploadWorkersCount,
	}

	for workerId := 0; workerId < uploadWorkersCount; workerId++ {
		workerCh := make(chan *PutFileCmd, 8)
		uploader.workers[workerId] = workerCh
		go uploader.worker(workerCh, workerId)
	}

	return uploader
}

func (uploader *FileUploader) UploadFile(cmd *PutFileCmd) {
	uploader.uploadFileMutex.Lock()
	defer uploader.uploadFileMutex.Unlock()
	uploader.uploadedFilesCounter += 1
	// calculate worker income queue
	workerId := uploader.uploadedFilesCounter % uint64(uploader.uploadWorkersCount)
	// put cmd to worker
	uploader.workers[workerId] <- cmd
}

func (uploader *FileUploader) worker(workerQ chan *PutFileCmd, workerId int) {
	for cmd := range workerQ {
		uploader.fileUploadWorker(cmd, workerId)
	}
}

func (uploader *FileUploader) fileUploadWorker(uploadFileCmd *PutFileCmd, workerId int) {
	now := time.Now()
	fileMeta := uploadFileCmd.FileMetadata
	fileData := uploadFileCmd.FileData
	shardGroup := sync.WaitGroup{}
	shardGroup.Add(len(fileMeta.Chunks))
	var errMsg error
	slog.Info("Start upload", "file", fileMeta.Id.String(), "workerId", workerId)

	for _, chunkInfo := range fileMeta.Chunks {
		go func(chunk *mosaicdb.ChunkInfo) {
			defer shardGroup.Done()
			chunkData := splitFile(fileData, chunk.Offset, chunk.Fsize)
			chunk.Id = utils.HashFrom(chunkData)
			slog.Info("Start upload chunk", "id", chunk.Id.String(), "fId", fileMeta.Id.String(), "shardId",
				chunk.ShardId, "size", chunk.Fsize)
			if err := uploader.engine.storage.Put(chunk.ShardId, chunk.Id, chunkData); err != nil {
				slog.Error("Err upload file", "id", fileMeta.Id.String(), "chunkId",
					chunk.Id.String(), "shardId", chunk.ShardId, "err", err.Error())
				errMsg = err
			}
		}(chunkInfo)
	}
	// wait all threads
	shardGroup.Wait()

	if errMsg == nil {
		fileMeta.Status = mosaicdb.FileStatusReady
		// remove tmp file
		uploader.engine.removeTempFile(fileMeta.Id)
	} else {
		fileMeta.Status = mosaicdb.FileStatusErr
	}
	// update file meta info
	uploader.engine.db.SaveFileMetadata(fileMeta)
	// try resolve upload err
	if errMsg != nil {
		uploader.engine.PrepareUploadErrCmd(fileMeta, errMsg)
	}
	slog.Info("End upload", "file", fileMeta.Id.String(), "size", fileMeta.FileSize, "took(ms)",
		time.Since(now).Milliseconds())
}

func splitFile(fileData []byte, offset uint64, size uint64) []byte {
	return fileData[offset : offset+size]
}
