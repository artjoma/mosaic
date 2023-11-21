# prepare folders
mkdir -p download
# get files after rebalancing 1
mkdir -p download/state_1

#############################################
# state 0
#############################################
# Add shards
# Return errors, already registered
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6751"}' | jq
curl http://0.0.0.0:25010/v1/shard/add -s -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6752"}' | jq

# Add new shard. Show state of cluster after rebalancing
curl -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6758"}' http://0.0.0.0:25010/v1/shard/add
sleep 1
curl -s http://0.0.0.0:25010/v1/cluster/state | jq
## Get meta information about file
curl -s http://0.0.0.0:25010/v1/file/meta/c122c61f1849b9a3dd688ea7f09a465e50dbbb3edb07b6644a812b907bae7996 | jq
curl -s http://0.0.0.0:25010/v1/file/meta/50d0d05a7e32d205cba982e0d83d4b511ccf53b7b812d68991a638a745fee98d | jq
curl -s http://0.0.0.0:25010/v1/file/meta/8e6f5f1ba3d39ce3e45f0a23e2c559963db042787fc09b12485872c80ec71958 | jq
curl -s http://0.0.0.0:25010/v1/file/meta/ce8ba53493edfe0d6186bc3f45544236fe82c421ff0afa377280123d4772c99c | jq

# Add new shard. Show state of cluster after rebalancing
curl -X POST -H "Content-Type: application/json" -d '{"host":"0.0.0.0:6759"}' http://0.0.0.0:25010/v1/shard/add | jq
# upload file when rebalancing process works
# b6dc92992515083e4923ee44e29757b8b1b175715d284b790b6f2e1b01440137
curl http://0.0.0.0:25010/v1/file/put -s -X POST -F file=@asset/rocket.jpeg | jq
sleep 1
curl -s http://0.0.0.0:25010/v1/cluster/state | jq
# Get meta information about file
curl -s http://0.0.0.0:25010/v1/file/meta/c122c61f1849b9a3dd688ea7f09a465e50dbbb3edb07b6644a812b907bae7996 | jq
curl -s http://0.0.0.0:25010/v1/file/meta/50d0d05a7e32d205cba982e0d83d4b511ccf53b7b812d68991a638a745fee98d | jq
curl -s http://0.0.0.0:25010/v1/file/meta/8e6f5f1ba3d39ce3e45f0a23e2c559963db042787fc09b12485872c80ec71958 | jq
curl -s http://0.0.0.0:25010/v1/file/meta/ce8ba53493edfe0d6186bc3f45544236fe82c421ff0afa377280123d4772c99c | jq
curl -s http://0.0.0.0:25010/v1/file/meta/b6dc92992515083e4923ee44e29757b8b1b175715d284b790b6f2e1b01440137 | jq

# Download after rebalancing files
wget -q http://0.0.0.0:25010/v1/file/download/c122c61f1849b9a3dd688ea7f09a465e50dbbb3edb07b6644a812b907bae7996 -O download/state_1/alarm.mp3
wget -q http://0.0.0.0:25010/v1/file/download/50d0d05a7e32d205cba982e0d83d4b511ccf53b7b812d68991a638a745fee98d -O download/state_1/car.png
wget -q http://0.0.0.0:25010/v1/file/download/8e6f5f1ba3d39ce3e45f0a23e2c559963db042787fc09b12485872c80ec71958 -O download/state_1/plant_catalog.xml
wget -q http://0.0.0.0:25010/v1/file/download/ce8ba53493edfe0d6186bc3f45544236fe82c421ff0afa377280123d4772c99c -O download/state_1/sun.jpg
wget -q http://0.0.0.0:25010/v1/file/download/b6dc92992515083e4923ee44e29757b8b1b175715d284b790b6f2e1b01440137 -O download/state_1/rocket.jpg