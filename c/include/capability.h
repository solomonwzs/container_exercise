#ifndef _MY_CAPABILITY_H
#define _MY_CAPABILITY_H

#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif

#include "base.h"
#include <netinet/in.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/capability.h>
#include <unistd.h>

#define list_caps do { \
  cap_t __caps = cap_get_proc(); \
  char *__txt = cap_to_text(__caps, NULL); \
  ldebug("%s\n", __txt); \
  cap_free(__txt); \
  cap_free(__caps); \
} while (0)

#define clear_all_caps do { \
  cap_t __caps = cap_init() \
} while (0)

#define check_cap_net_raw do {\
  int __sock; \
  if ((__sock = socket(AF_INET6, SOCK_RAW, IPPROTO_RAW)) < 0) {\
    lperror("cap_net_raw"); \
  } else { \
    ldebug("cap_net_raw ok\n"); \
    close(__sock); \
  } \
} while (0)

#endif
