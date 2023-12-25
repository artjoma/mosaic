golangci-lint run

go build -ldflags "-s -w" -o build/mosaic
#env GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o build/mosaic.exe
#env GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o build/mosaic-darwin

# use UPX compressor
if [ ! -z $1 ]; then
   upx build/mosaic
fi