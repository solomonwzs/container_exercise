.SUFFIXES: .c

C		= gcc
AR		= ar
CFLAGS	= -Wall -fpic -g -c -I./include -I./deps/iproute2/include
C_PATH	= "csrc"
C_SRC	= $(wildcard csrc/*.c)

OBJ			= $(C_SRC:%.c=%.o)
DEP			= $(OBJ:%.o=%.d)

BASE	= ${PWD}
GOPATH	:= ${BASE}:${GOPATH}
GO		?= go
GO_SRC	= $(shell find ./src -iname *.go)

IPROUTE2_PATH 	= "./deps/iproute2"


./build/main.goc:$(GO_SRC) ./build ./build/libnetwork.a ./src/ctype.go
	@echo -e "\033[0;33m*\033[0m $@"
	@$(GO) build -o $@ ./src/*.go

-include $(DEP)

.c.o:
	@echo -e "\033[0;33m*\033[0m $@"
	@$(C) $(CFLAGS) -MMD $< -o $@

./build:
	@echo -e "\033[0;33m*\033[0m $@"
	@[ -d $@ ] || mkdir $@

./deps/iproute2/lib/*.a:
	@echo -e "\033[0;33m*\033[0m $@"
	@git submodule init
	@git submodule update
	@cd $(IPROUTE2_PATH) && ./configure
	@make -C $(IPROUTE2_PATH)

./build/libnetwork.a:./deps/iproute2/lib/*.a ./build $(OBJ)
	@echo -e "\033[0;33m*\033[0m $@"
	@$(AR) -rs $@ $(OBJ) $(IPROUTE2_PATH)/ip/rtm_map.o

./src/ctype.go:./include/ctype.go ./include/cont_proto.h
	@echo -e "\033[0;33m*\033[0m $@"
	@$(GO) tool cgo -godefs $< > $@

.PHONY: clean
clean:
	@rm $(OBJ) $(DEP)
	@[ -d ./build ] && rm -r ./build
