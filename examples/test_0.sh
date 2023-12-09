set -x

# prepare folders
mkdir -p download
# download files after fresh setup
mkdir -p download/state_0

#############################################
# state 0
#############################################
# Add shards
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6751"}' | jq
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6752"}' | jq
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6753"}' | jq
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6754"}' | jq
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6755"}' | jq
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6756"}' | jq
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6757"}' | jq

# Add files
# h256: c122c61f1849b9a3dd688ea7f09a465e50dbbb3edb07b6644a812b907bae7996
curl http://0.0.0.0:25010/v1/file/put -s -X POST -F file=@asset/alarm.mp3 | jq
curl -s http://0.0.0.0:25010/v1/cluster/state | jq

# h256:50d0d05a7e32d205cba982e0d83d4b511ccf53b7b812d68991a638a745fee98d
curl http://0.0.0.0:25010/v1/file/put -s -X POST -F file=@asset/car.png | jq
curl -s http://0.0.0.0:25010/v1/cluster/state | jq

# h256:8e6f5f1ba3d39ce3e45f0a23e2c559963db042787fc09b12485872c80ec71958
curl http://0.0.0.0:25010/v1/file/put -s -X POST -F file=@asset/plant_catalog.xml | jq
curl -s http://0.0.0.0:25010/v1/cluster/state | jq

# h256:ce8ba53493edfe0d6186bc3f45544236fe82c421ff0afa377280123d4772c99c
curl http://0.0.0.0:25010/v1/file/put -s -X POST -F file=@asset/sun.jpg | jq
curl -s http://0.0.0.0:25010/v1/cluster/state | jq

# Get meta information about file
curl -s http://0.0.0.0:25010/v1/file/meta/c122c61f1849b9a3dd688ea7f09a465e50dbbb3edb07b6644a812b907bae7996 | jq
curl -s http://0.0.0.0:25010/v1/file/meta/50d0d05a7e32d205cba982e0d83d4b511ccf53b7b812d68991a638a745fee98d | jq
curl -s http://0.0.0.0:25010/v1/file/meta/8e6f5f1ba3d39ce3e45f0a23e2c559963db042787fc09b12485872c80ec71958 | jq
curl -s http://0.0.0.0:25010/v1/file/meta/ce8ba53493edfe0d6186bc3f45544236fe82c421ff0afa377280123d4772c99c | jq

sleep 1

# Download files
wget -q http://0.0.0.0:25010/v1/file/download/c122c61f1849b9a3dd688ea7f09a465e50dbbb3edb07b6644a812b907bae7996 -O download/state_0/alarm.mp3
wget -q http://0.0.0.0:25010/v1/file/download/50d0d05a7e32d205cba982e0d83d4b511ccf53b7b812d68991a638a745fee98d -O download/state_0/car.png
wget -q http://0.0.0.0:25010/v1/file/download/8e6f5f1ba3d39ce3e45f0a23e2c559963db042787fc09b12485872c80ec71958 -O download/state_0/plant_catalog.xml
wget -q http://0.0.0.0:25010/v1/file/download/ce8ba53493edfe0d6186bc3f45544236fe82c421ff0afa377280123d4772c99c -O download/state_0/sun.jpg
