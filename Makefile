
build:
	go build -o $$GOPATH/bin/download-manager -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH;
