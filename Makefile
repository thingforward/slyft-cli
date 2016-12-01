
ARCHUNIX=amd64
ARCHWIN=386

OSX=darwin
LINUX=linux
WIN=windows

all: slyft.exe slyft.mac slyft.lin


slyft.exe:
	GOARCH=$(ARCHWIN) GOOS=$(WIN) go build -o slyft.exe *.go

slyft.mac:
	GOARCH=$(ARCHUNIX) GOOS=$(OSX) go build -o slyft.mac *.go

slyft.lin:
	GOARCH=$(ARCHUNIX) GOOS=$(LINUX) go build -o slyft.lin *.go


slyft: main.go user_manager.go *.go
	go build -o slyft *.go

