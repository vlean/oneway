
allbuild: build webuild

## 构建
webuild:
	mkdir -p bin
	cd fe && npm run build && mv dist ../bin/

build:
	mkdir -p bin
	cp config.toml bin/config.toml
	go build -tags "nethttpomithttp2" -o bin/oneway *.go

## 启动
runserver:
	cd bin && ./oneway server -c config.toml
runclient:
	cd bin && ./oneway client -c config.toml
runweb:
	cd fe && npm run dev

server: build runserver

client: build runclient

web: runweb
