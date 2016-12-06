
ARCHUNIX=amd64
ARCHWIN=386

OSX=darwin
LINUX=linux
WIN=windows

.DEFAULT_GOAL := slyft

all: slyft.exe slyft.mac slyft.lin

slyft.exe:
	GOARCH=$(ARCHWIN) GOOS=$(WIN) go build -o slyft.exe *.go

slyft.mac:
	GOARCH=$(ARCHUNIX) GOOS=$(OSX) go build -o slyft.mac *.go

slyft.lin:
	GOARCH=$(ARCHUNIX) GOOS=$(LINUX) go build -o slyft.lin *.go

slyft: main.go *.go
	go build -o slyft *.go

clean:
	rm -f slyft.exe slyft.mac slyft.lin slyft
