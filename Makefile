BINNAME=refactor
RELEASE=-s -w
UPXBIN=/usr/local/bin/upx
GOBIN=/usr/local/bin/go
GOOS=$(shell uname -s | tr [A-Z] [a-z])
GOARGS=GOARCH=amd64 CGO_ENABLED=1
GOBUILD=$(GOARGS) $(GOBIN) build -ldflags="$(RELEASE)"

.PHONY: all build clean upx upxx

all: clean build
build:
	@echo "Compile $(BINNAME) ..."
	GOOS=$(GOOS) $(GOBUILD) -o $(BINNAME) ./cmd/reverse/
	@echo "Build success."
clean:
	rm -f $(BINNAME)
	@echo "Clean all."
upx: build command
	$(UPXBIN) $(BINNAME)
upxx: build command
	$(UPXBIN) --ultra-brute $(BINNAME)
