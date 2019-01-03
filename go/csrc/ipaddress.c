#include "network.h"
#include "utils.h"
#include <string.h>

#ifndef INFINITY_LIFE_TIME
#define INFINITY_LIFE_TIME 0xFFFFFFFFU
#endif


struct ipaddr_req {
  struct nlmsghdr   n;
  struct ifaddrmsg  ifa;
  char              buf[256];
};


static inline void
init_ipaddr_req(struct ipaddr_req *req, uint16_t type, uint16_t flags) {
  memset(req, 0, sizeof(struct ipaddr_req));
  req->n.nlmsg_len = NLMSG_LENGTH(sizeof(struct ifaddrmsg));
  req->n.nlmsg_flags = NLM_F_REQUEST | flags;
  req->n.nlmsg_type = type;
  req->ifa.ifa_family = 0;
}


static int
default_scope(inet_prefix *local) {
  if (local->family == AF_INET) {
    if (local->bytelen >= 1 && *(uint8_t *)&local->data == 127) {
      return RT_SCOPE_HOST;
    }
  }
  return 0;
}


static bool
ipaddr_is_multicast(inet_prefix *addr) {
  if (addr->family == AF_INET) {
    return IN_MULTICAST(ntohl(addr->data[0]));
  } else if (addr->family == AF_INET6) {
    return IN6_IS_ADDR_MULTICAST(addr->data);
  } else {
    return false;
  }
}


static int
ipaddr_modify(uint16_t msg_type, uint32_t flags, int argc, char **argv) {
  struct ipaddr_req req;
  init_ipaddr_req(&req, msg_type, flags);

  char *dev = NULL;

  inet_prefix local;
  int local_len = 0;

  inet_prefix peer;
  int peer_len = 0;

  uint32_t ifa_flags = 0;
  int brd_len = 0;

  int scoped = 0;

  while (argc > 0) {
    if (matches(*argv, "broadcast") == 0 ||
        strcmp(*argv, "brd") == 0) {
      _NEXT_ARG;
      if (brd_len) {
        return -1;
      }

      if (strcmp(*argv, "+") == 0) {
        brd_len = -1;
      } else if (strcmp(*argv, "-") == 0) {
        brd_len = -2;
      } else {
        inet_prefix addr;
        get_addr(&addr, *argv, req.ifa.ifa_family);
        if (req.ifa.ifa_family == AF_UNSPEC) {
          req.ifa.ifa_family = addr.family;
        }
        addattr_l(&req.n, sizeof(req), IFA_BROADCAST, &addr.data,
                  addr.bytelen);
        brd_len = addr.bytelen;
      }
    } else if (strcmp(*argv, "dev") == 0) {
      _NEXT_ARG;
      dev = *argv;
    } else {
      if (strcmp(*argv, "local") == 0) {
        _NEXT_ARG;
      }

      if (local_len) {
        return -1;
      }

      get_prefix(&local, *argv, req.ifa.ifa_family);
      if (req.ifa.ifa_family == AF_UNSPEC) {
        req.ifa.ifa_family = local.family;
      }
      addattr_l(&req.n, sizeof(req), IFA_LOCAL, &local.data, local.bytelen);
      local_len = local.bytelen;
    }
    _NEXT_ARG;
  }

  if (ifa_flags <= 0xff) {
    req.ifa.ifa_flags = ifa_flags;
  } else {
    addattr32(&req.n, sizeof(req), IFA_FLAGS, ifa_flags);
  }

  if (dev == NULL) {
    return -1;
  }

  if (peer_len == 0 && local_len) {
    if (msg_type == RTM_DELADDR && local.family == AF_INET &&
        !(local.flags & PREFIXLEN_SPECIFIED)) {
      return -1;
    } else {
      peer = local;
      addattr_l(&req.n, sizeof(req), IFA_ADDRESS, &local.data,
                local.bytelen);
    }
  }

  if (req.ifa.ifa_prefixlen == 0) {
    req.ifa.ifa_prefixlen = local.bitlen;
  }

  if (brd_len < 0 && msg_type != RTM_DELADDR) {
    if (req.ifa.ifa_family != AF_INET) {
      return -1;
    }

    inet_prefix brd = peer;
    if (brd.bitlen <= 30) {
      for (int i = 31; i >= brd.bitlen; i--) {
        if (brd_len == -1) {
          brd.data[0] |= htonl(1<<(31-i));
        } else {
          brd.data[0] &= ~htonl(1<<(31-i));
        }
      }
      addattr_l(&req.n, sizeof(req), IFA_BROADCAST, &brd.data, brd.bytelen);
      brd_len = brd.bytelen;
    }
  }

  if (!scoped && msg_type != RTM_DELADDR) {
    req.ifa.ifa_scope = default_scope(&local);
  }

  req.ifa.ifa_index = ll_name_to_index(dev);
  if (!req.ifa.ifa_index) {
    return -1;
  }

  if ((ifa_flags & IFA_F_MCAUTOJOIN) && !ipaddr_is_multicast(&local)) {
    return -1;
  }

  return send_rtnl_message(&req.n);
}


int
ipaddr_add(int argc, char **argv) {
  return ipaddr_modify(RTM_NEWADDR, NLM_F_CREATE|NLM_F_EXCL, argc, argv);
}
