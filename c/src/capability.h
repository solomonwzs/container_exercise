#ifndef _MY_CAPABILITY_H
#define _MY_CAPABILITY_H

#include <sys/capability.h>

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

#endif
