
allbuild: webuild build

## 构建
webuild:
	mkdir -p bin && rm -rf bin/dist
	cd fe && npm run build && mv dist ../bin/

build:
	mkdir -p bin
	cp config.toml bin/config.toml
	GOOS=linux GARCH=amd64 go build -tags "nethttpomithttp2" -o bin/oneway *.go

pull:
	git pull
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

dbuild:
	rm -rf bin/app && mkdir -p bin/app 
	cp bin/oneway bin/app/
	cp example.toml bin/app/config.toml
	docker build -t oneway .

alldb: allbuild dbuild

clean:
	docker ps -a|grep -v CONTAINER |awk '{print $1}'|xargs -i docker rm {}
	docker images|grep -v REPOSITORY|awk '{print $3}'|xargs -i docker rmi {}