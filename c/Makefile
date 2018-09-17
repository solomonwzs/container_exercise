.SUFFIXES: .c

C			= gcc
CFLAGS		= -Wall -fpic -g -c
ALL_SRC		= $(wildcard src/*.c)

MAIN_PATH		= src
BUILD_PATH		= bin
BASESYS_PATH	= /home/solomon/workspace/basesystem/debian/

MACROS	=


.c.o:
	@echo -e "\e[1;3;33m*\e[0m $@"
	@$(C) $(CFLAGS) $(MACROS) $< -o $@

all:obj pre
	@$(C) $(MAIN_PATH)/*.o -o $(BUILD_PATH)/main
	@echo -e "\e[1;3;32m[OK]\e[0m main ok"

obj:$(ALL_SRC:%.c=%.o)

pre:
	@[ -d "$(BUILD_PATH)" ] || mkdir -p $(BUILD_PATH)

test: all
	@$(BUILD_PATH)/main "$(BASESYS_PATH)"
