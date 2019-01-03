#include "network.h"


int
send_rtnl_message(struct nlmsghdr *n) {
  struct rtnl_handle rth = {.fd = -1};
  if (rtnl_open(&rth, 0) < 0) {
    return -1;
  }
  if (rtnl_talk(&rth, n, NULL) < 0) {
    rtnl_close(&rth);
    return -1;
  }
  rtnl_close(&rth);
  return 0;
}
