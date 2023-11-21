package engine

import (
	"bytes"
	"fmt"
	"github.com/pierrec/lz4/v4"
	"github.com/zeebo/blake3"
	"io"
	"log/slog"
	"mime/multipart"
	"mosaic/mosaicdb"
	"mosaic/types"
	"mosaic/utils"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const LZ4CmpLevel = lz4.Level7

func (engine *Engine) putFileCmd(cmd *PutFileCmd) error {
	fileMeta := cmd.FileMetadata
	slog.Info("Put", "file", fileMeta.Id.String(), "size", fileMeta.FileSize, "oSize", fileMeta.OriginalFileSize)
	shardsSize := engine.shardsSize()
	slog.Info("Shards", "state", shardsSize)
	fileDataChunks := splitFileToChunks(fileMeta.FileSize, shardsSize)
	slog.Info("File chunks", "shards", fileDataChunks)
	// update shards size
	engine.db.UpdateShardsSize(fileDataChunks, true)
	// prepare info
	fileMeta.Chunks = fileChunkToChunkInfo(fileDataChunks)

	strBuff := bytes.NewBufferString("")
	for _, chunk := range fileMeta.Chunks {
		strBuff.WriteString(fmt.Sprintf("{sId:%d offset:%d size:%d}", chunk.ShardId, chunk.Offset, chunk.Fsize))
	}

	slog.Info("Chunk", "info", strBuff.String())
	fileMeta.Status = mosaicdb.FileStatusUploading
	// update info with status Uploading + chunks info
	if err := engine.db.SaveFileMetadata(fileMeta); err != nil {
		return err
	}
	// upload file
	engine.fileUploader.UploadFile(cmd)

	return nil
}

// PreparePutFileCmd put file -> proces file
func (engine *Engine) PreparePutFileCmd(f multipart.File, originalFName string) (*mosaicdb.FileMetadata, error) {
	// we can't lose file
	tempFilePath := engine.prepareTempFile()
	outFile, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	// TODO Get hash when dump file to disk !
	// write uploaded file to temp file using LZ4
	originalFileSize, fileHash, err := engine.compressFile(f, outFile)
	if err != nil {
		slog.Error("Failed to compress file", "err", err.Error())
		return nil, err
	}

	// check if file already uploaded
	checkMeta, err := engine.GetFileMetadata(fileHash)
	if checkMeta != nil {
		slog.Info("File already exists", "fId", fileHash.String(), "oSize", checkMeta.OriginalFileSize)
		// remove temp file
		os.Remove(tempFilePath)
		return checkMeta, nil
	}

	// rename temp file
	fileName := filepath.Join(engine.tempFolder, fileHash.String())
	if err := os.Rename(tempFilePath, fileName); err != nil {
		slog.Error("Failed to rename", "file", tempFilePath)
		return nil, err
	}

	// read compressed file to mem
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	fMeta := &mosaicdb.FileMetadata{
		Id:               fileHash,
		FileSize:         uint64(len(fileData)),
		OriginalFileSize: originalFileSize,
		OriginalFileName: originalFName,
		Status:           mosaicdb.FileStatusPending,
	}
	// register new file with status Pending
	if err := engine.db.SaveFileMetadata(fMeta); err != nil {
		return nil, err
	}
	// we should store files on disk to prevent file loss
	engine.copyToTempFile(fMeta.Id, fileData)
	// put to coordinator queue, process new files async
	engine.commandChan <- &PutFileCmd{FileMetadata: fMeta, FileData: fileData}
	// unblock API request
	return fMeta, nil
}

func (engine *Engine) prepareTempFile() string {
	now := time.Now()
	randomStr := utils.RandomString(5)
	fileName := "tmp_" + strconv.Itoa(int(now.Unix())) + "_" + randomStr
	slog.Info("API start put", "file", fileName)
	return filepath.Join(engine.tempFolder, fileName)
}

func (engine *Engine) copyToTempFile(fileId types.H256, data []byte) {
	fPath := filepath.Join(engine.tempFolder, fileId.String())
	err := os.WriteFile(fPath, data, 0644)
	if err != nil {
		slog.Error("Failed to write file", "err", err.Error())
	}
}

// Compress file using LZ4 algo.
// Return bytes written, H256 of original file, error
func (engine *Engine) compressFile(in multipart.File, outFile *os.File) (int64, types.H256, error) {
	lzWriter := lz4.NewWriter(outFile)
	hasher := blake3.New()
	lzWriter.Apply(lz4.CompressionLevelOption(lz4.Level8))
	defer lzWriter.Close()
	n, err := io.Copy(io.MultiWriter(hasher, lzWriter), in)
	if err != nil {
		return 0, nil, err
	}

	return n, hasher.Sum(nil), nil
}

func fileChunkToChunkInfo(fileDataChunks map[types.ShardId]uint64) []*mosaicdb.ChunkInfo {
	chunks := make([]*mosaicdb.ChunkInfo, 0, len(fileDataChunks))
	offset := uint64(0)
	// prepare chunks info
	for id, size := range fileDataChunks {
		if size == 0 {
			continue
		}
		// Chunk id will be calculated when chunk will be ready to upload
		chunk := &mosaicdb.ChunkInfo{
			Fsize:   size,
			Offset:  offset,
			ShardId: id,
		}
		chunks = append(chunks, chunk)
		offset += size
	}

	return chunks
}
