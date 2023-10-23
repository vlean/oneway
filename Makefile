
build:
	mkdir -p bin
	go build -tags "nethttpomithttp2" -o bin/oneway *.go

server: build
	cp config.toml bin/config.toml
	cd bin && ./oneway server -c config.toml

client: build
	cp config.toml bin/config.toml
	cd bin && ./oneway client -c config.toml
