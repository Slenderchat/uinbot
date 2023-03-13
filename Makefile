all: uinbot uin rosreestr config
uinbot: uinbot.go
	go build uinbot.go
config: config/config.go
	pushd config; \
	go build -buildmode=plugin config.go; \
	mv config.so ..; \
	popd
uin: uin/uin.go
	pushd uin; \
	go build -buildmode=plugin uin.go; \
	mv uin.so ..; \
	popd
rosreestr: rosreestr/rosreestr.go
	pushd rosreestr; \
	go build -buildmode=plugin rosreestr.go; \
	mv rosreestr.so ..; \
	popd