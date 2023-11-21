package types

import (
	"encoding/hex"
	"encoding/json"
)

/*
	Global types
*/

type AppConfig struct {
	ApiHttpPort              uint16 `env:"API_HTTP_PORT" envDefault:"25010"`
	ApiHttpAddress           string `env:"API_HTTP_ADDRESS" envDefault:"0.0.0.0"`
	FilesTempFolder          string `env:"FILES_TEMP_FOLDER" envDefault:"/tmp/masaic/temp"` // http file upload folder
	DbPath                   string `env:"API_DB_PATH" envDefault:""`
	FileUploadWorkersCount   uint8  `env:"FILE_UPLOAD_WORKERS_COUNT" envDefault:"10"`
	FileDownloadWorkersCount uint8  `env:"FILE_DOWNLOAD_WORKERS_COUNT" envDefault:"10"`
}

// H256 hash 256 bit length
type H256 []byte

func (h H256) String() string {
	return hex.EncodeToString(h)
}

func (h H256) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

type ShardId uint32
