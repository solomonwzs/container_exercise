#ifndef BASE_H
#define BASE_H

#include <errno.h>
#include <stdio.h>
#include <unistd.h>

#define COLOR_ERR "\e[2;3;31m"
#define COLOR_WAR "\e[2;3;33m"
#define COLOR_DEB "\e[2;3;32m"
#define COLOR_INF "\e[2;3;37m"

#define _print(_fmt_, _color_, ...) \
    fprintf(stderr, _color_ "=%d= [%s:%d:%s]\e[0m " _fmt_, \
            getpid(), __FILE__, __LINE__, __func__, ## __VA_ARGS__)

#define ldebug(_fmt_, ...) _print(_fmt_, COLOR_DEB, ## __VA_ARGS__)

#define lperror(_s_) _print("%s: %s\n", COLOR_ERR, _s_, strerror(errno))

#endif
