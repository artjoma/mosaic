package engine

import (
	"mosaic/mosaicdb"
	"mosaic/types"
)

type ProcessorCommand interface {
	getCmdId() uint8
}

type PutFileCmd struct {
	FileMetadata *mosaicdb.FileMetadata
	FileData     []byte
}

func (cmd *PutFileCmd) getCmdId() uint8 {
	return 1
}

type AddShardCmd struct {
	Host string
	Id   types.ShardId
}

func (cmd *AddShardCmd) getCmdId() uint8 {
	return 2
}

type FileUploadErrCmd struct {
	FileMetadata *mosaicdb.FileMetadata
	Err          error
}

func (cmd *FileUploadErrCmd) getCmdId() uint8 {
	return 3
}
