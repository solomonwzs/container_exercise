.SUFFIXES: .c

C		= gcc
AR		= ar
CFLAGS	= -Wall -fpic -g -c -I./include -I./deps/iproute2/include
C_PATH	= "csrc"
C_SRC	= $(wildcard csrc/*.c)

BASE	= ${PWD}
GOPATH	:= ${BASE}:${GOPATH}
GO		?= go
GO_SRC	= $(shell find ./src -iname *.go)

IPROUTE2_PATH 	= "./deps/iproute2"


./build/main.goc:$(GO_SRC) ./build ./build/libnetwork.a ./src/ctype.go
	@$(GO) build -o $@ ./src/*.go

.c.o:
	@$(C) $(CFLAGS) $< -o $@

./build:
	@[ -d $@ ] || mkdir $@

./deps/iproute2/lib/*.a:
	@git submodule update
	@cd $(IPROUTE2_PATH) && ./configure
	@make -C $(IPROUTE2_PATH)

./build/libnetwork.a:./deps/iproute2/lib/*.a ./build $(C_SRC:%.c=%.o)
	@$(AR) -rs $@ $(C_PATH)/*.o $(IPROUTE2_PATH)/ip/rtm_map.o

./src/ctype.go:./include/ctype.go ./include/cont_proto.h
	@$(GO) tool cgo -godefs $< > $@

.PHONY: clean
clean:
	@rm $(C_PATH)/*.o
	@rm -r ./build
