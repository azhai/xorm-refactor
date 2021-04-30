BINNAME=refactor
RELEASE=-s -w

GOOS=$(shell uname -s | tr [A-Z] [a-z])
GOARGS=GOARCH=amd64 CGO_ENABLED=1
ifeq ($(GOOS),windows)
    GOBIN=go
    UPXBIN=
else
    ifeq ($(GOOS),darwin)
        GOBIN=/usr/local/bin/go
        UPXBIN=/usr/local/bin/upx
    else
        GOBIN=/usr/bin/go
        UPXBIN=/usr/bin/upx
    endif
endif
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
