
build:
	mkdir -p bin
	go build -tags "nethttpomithttp2" -o bin/oneway *.go

pull:
	git pull 

server: build
	cp config.toml bin/config.toml
	cd bin && ./oneway server -c config.toml

client: build
	cp config.toml bin/config.toml
	cd bin && ./oneway client -c config.toml
