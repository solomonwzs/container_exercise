#include "base.h"
#include <sys/capability.h>

void list_caps() {
  cap_t caps = cap_get_proc();
  ssize_t s = 0;
  ldebug("%s\n", cap_to_text(caps, &s));
  cap_free(caps);
}
