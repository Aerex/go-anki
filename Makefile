VERCMD  ?= git describe --long --tags 2> /dev/null
VERSION ?= $(shell $(VERCMD) || cat VERSION)
BINNAME ?= "anki"
OUT 		?= $(pwd)/build

PREFIX    ?= /usr/local

.PHONY: all build clean

all: build 

build:
	go build -o $$(pwd)/build/$(BINNAME) 

clean:
	rm -rf $(OUT)
