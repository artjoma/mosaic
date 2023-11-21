package api

import (
	"mosaic/mosaicdb"
	"mosaic/types"
)

type ErrResponse struct {
	ErrMsg string `json:"errMsg"`
}

func NewErrResponse(err error) *ErrResponse {
	return &ErrResponse{
		ErrMsg: err.Error(),
	}
}

type AddShardRequest struct {
	Host string `json:"host"`
}

type AddShardResponse struct {
	ShardId types.ShardId `json:"shardId"`
}

type PutFileResponse struct {
	FileMetadata *mosaicdb.FileMetadata `json:"fileMetadata"`
}
