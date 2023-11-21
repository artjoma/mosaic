package engine

import (
	"bytes"
	"fmt"
	"github.com/pierrec/lz4/v4"
	"io"
	"log/slog"
	"mosaic/mosaicdb"
	"mosaic/storage"
	"mosaic/utils"
	"os"
	"testing"
	"time"
)

func TestFileUploader(t *testing.T) {
	database := mosaicdb.NewDatabase("")
	storage := storage.NewStorage()
	engine, err := NewEngine("/tmp/masaic/temp", database, storage, 10, 10)
	if err != nil {
		panic(err)
	}
	host := fmt.Sprintf("0.0.0.0:%d", utils.TestPortRangeFrom)
	engine.addShardsCmd(&AddShardCmd{
		Host: host,
		Id:   mosaicdb.BuildShardId(host),
	})
	host = fmt.Sprintf("0.0.0.0:%d", utils.TestPortRangeFrom+1)
	engine.addShardsCmd(&AddShardCmd{
		Host: host,
		Id:   mosaicdb.BuildShardId(host),
	})

	fileData := utils.RandomBytes(1024 * 1024)
	filePath := "/tmp/test.f"
	os.WriteFile(filePath, fileData, 0774)
	file, _ := os.Open(filePath)
	fileMeta, err := engine.PreparePutFileCmd(file, "some")
	fileId := fileMeta.Id
	slog.Info("File", "id", fileId.String())
	// TODO check DB when file status changed to ready
	time.Sleep(time.Second * 2)

	_, _fileData, err := engine.DownloadFile(fileId)
	if err != nil {
		t.Fatal(err.Error())
	}

	zr := lz4.NewReader(_fileData)
	if err := zr.Apply(lz4.ConcurrencyOption(4)); err != nil {
		t.Fatal(err)
	}
	fileUncp, err := io.ReadAll(zr)

	err = os.WriteFile("/tmp/test.pdf", fileUncp, 0774)

	if !bytes.Equal(fileData, fileUncp) {
		t.Error("Invalid content")
	}

}
