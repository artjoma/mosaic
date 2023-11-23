package engine

import (
	"mime/multipart"
	"mosaic/mosaicdb"
	"mosaic/types"
)

type ProcessorCommand interface {
	getCmdId() uint8
	Prepare(engine *Engine) error
	Execute() error
}

type PutFileCmd struct {
	engine *Engine
	// prepare
	file             multipart.File
	OriginalFileName string

	FileMetadata *mosaicdb.FileMetadata
	FileData     []byte
}

func (cmd *PutFileCmd) SetPrepare(f multipart.File, originalFName string) {
	cmd.file = f
	cmd.OriginalFileName = originalFName
}

func (cmd *PutFileCmd) getCmdId() uint8 {
	return 1
}

type AddShardCmd struct {
	engine *Engine
	Host   string
	Id     types.ShardId
}

func (cmd *AddShardCmd) getCmdId() uint8 {
	return 2
}

type FileUploadErrCmd struct {
	engine       *Engine
	FileMetadata *mosaicdb.FileMetadata
	Err          error
}

func (cmd *FileUploadErrCmd) getCmdId() uint8 {
	return 3
}
