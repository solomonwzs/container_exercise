.SUFFIXES: .c

C		= gcc
AR		= ar
CFLAGS	= -Wall -fpic -g -c -I./include -I./deps/iproute2/include
C_PATH	= "csrc"
C_SRC	= $(wildcard csrc/*.c)

BASE	= ${PWD}
GOPATH	:= ${BASE}:${GOPATH}
GO		?= go
GO_SRC	= "src"

BUILD_PATH = "build"

.c.o:
	@echo -e "\033[0;33m*\033[0m $@"
	@$(C) $(CFLAGS) $< -o $@

all:pre libnetwork
	@$(GO) build -o $(BUILD_PATH)/main.goc $(GO_SRC)/*.go

obj:$(C_SRC:%.c=%.o)

pre:
	@[ -d $(BUILD_PATH) ] || mkdir $(BUILD_PATH)

dep-iproute2:
	@cd ./deps/iproute2 && ./configure
	@make -C ./deps/iproute2

libnetwork:pre obj
	@$(AR) -rs $(BUILD_PATH)/libnetwork.a $(C_PATH)/*.o

clean:
	@rm $(C_PATH)/*.o
	@rm -r $(BUILD_PATH)
