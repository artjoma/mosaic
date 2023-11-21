# Mosaic

Mosaic is POC of distributed file storage.

### Key fetures
 * Add new shards on the flight. All files will be rebalanced.
 * Lz4 file compression.
 * Shard based on Apache KVRocks(RocksDB under the hood) and Redis protocol.
 * Internal storage based on embedded Pebble key value database. Information about file stored at MsgPack format.
 * API allow concurrently upload and download files.
 * File chunks id and file id based on Blake3s hash function.
 * HTTP API with common functionality.
### Limitation
 * Max file chunk per shard 512Mb

### HTTP API
 * v1/shard/add
```shell
  curl -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6751"}' http://0.0.0.0:25010/v1/shard/add
 ```
 * /file/put
```shell
 curl -F file=@OIPD.pdf http://0.0.0.0:25010/v1/file/put
 ```
 * /file/download/:fId
```shell
  curl http://0.0.0.0:25010/v1/file/download/<fileId>
 ```
 * /file/meta/:fId
```shell
  curl http://0.0.0.0:25010/v1/file/meta/:fileId
```
 * /cluster/state
```shell
    curl http://0.0.0.0:25010/v1/cluster/state
 ```

### How to play
```shell
# Setup shards
docker compose up+stop 
docker build -t mosaic .
docker run --network="host" -p 25010:25010 mosaic
# Multiple examples using HTTP API
cd examples
sh test_0.sh
sh test_1.sh
sh test_2.sh
```