build: build-config build-uin
	go build uinbot.go
build-config:
	pushd config; \
	go build -buildmode=plugin config.go; \
	mv config.so ..; \
	popd
build-uin:
	pushd uin; \
	go build -buildmode=plugin uin.go; \
	mv uin.so ..; \
	popd