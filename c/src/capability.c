#include "base.h"
#include "capability.h"
#include <sys/prctl.h>


void init_cap() {
  cap_value_t cap_values[] = {
    CAP_SETUID, CAP_SETGID, CAP_SETPCAP, CAP_NET_RAW};

  cap_t caps = cap_init();
  cap_set_flag(caps, CAP_PERMITTED, 4, cap_values, CAP_SET);
  cap_set_flag(caps, CAP_EFFECTIVE, 4, cap_values, CAP_SET);
  cap_set_proc(caps);
  prctl(PR_SET_KEEPCAPS, 1, 0, 0, 0);
  cap_free(caps);

  setgid(1000);
  setuid(1000);

  caps = cap_get_proc();
  cap_set_flag(caps, CAP_EFFECTIVE, 4, cap_values, CAP_SET);
  cap_set_flag(caps, CAP_INHERITABLE, 4, cap_values, CAP_SET);
  cap_set_proc(caps);
  cap_free(caps);

  list_caps;
}
