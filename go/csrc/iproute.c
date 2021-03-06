#include "network.h"
#include "utils.h"
#include <string.h>


struct iproute_req {
  struct nlmsghdr	n;
  struct rtmsg		r;
  char			      buf[4096];
};


static inline void
init_iproute_req(struct iproute_req *req, uint16_t type, uint16_t flags) {
  memset(req, 0, sizeof(struct iproute_req));
  req->n.nlmsg_len = NLMSG_LENGTH(sizeof(struct rtmsg));
  req->n.nlmsg_flags = NLM_F_REQUEST | flags;
  req->n.nlmsg_type = type;
  req->r.rtm_family = 0;
  req->r.rtm_table = RT_TABLE_MAIN;
  req->r.rtm_scope = RT_SCOPE_NOWHERE;

  if (type != RTM_DELROUTE) {
    req->r.rtm_protocol = RTPROT_BOOT;
    req->r.rtm_scope = RT_SCOPE_UNIVERSE;
    req->r.rtm_type = RTN_UNICAST;
  }
}


static int
iproute_modify(uint16_t msg_type, uint32_t flags, int argc, char **argv) {
  struct iproute_req req;
  init_iproute_req(&req, msg_type, flags);

  char mxbuf[256];
  struct rtattr *mxrta = (void *)mxbuf;
  mxrta->rta_type = RTA_METRICS;
  mxrta->rta_len = RTA_LENGTH(0);

  int dst_ok = 0;
  int gw_ok = 0;
  int nhs_ok = 0;
  int scope_ok = 0;
  int table_ok = 0;
  int type_ok = 0;

  char *d = NULL;

  while (argc > 0) {
    if (strcmp(*argv, "via") == 0) {
      if (gw_ok) {
        return -1;
      }
      gw_ok = 1;
      _NEXT_ARG;

      int family = read_family(*argv);
      if (family == AF_UNSPEC) {
        family = req.r.rtm_family;
      } else {
        _NEXT_ARG;
      }

      inet_prefix addr;
      get_addr(&addr, *argv, family);
      if (req.r.rtm_family == AF_UNSPEC) {
        req.r.rtm_family = addr.family;
      }
      if (addr.family == req.r.rtm_family) {
        addattr_l(&req.n, sizeof(req), RTA_GATEWAY, &addr.data,
                  addr.bytelen);
      } else {
        addattr_l(&req.n, sizeof(req), RTA_VIA, &addr.family,
                  addr.bytelen + 2);
      }
    } else if (strcmp(*argv, "dev") == 0 ||
               strcmp(*argv, "oif") == 0) {
      _NEXT_ARG;
      d = *argv;
    } else {
      if (strcmp(*argv, "to") == 0) {
        _NEXT_ARG;
      }

      int type;
      if ((**argv < '0' || **argv > '9') &&
          rtnl_rtntype_a2n(&type, *argv) == 0) {
        _NEXT_ARG;
        req.r.rtm_type = type;
        type_ok = 1;
      }

      if (dst_ok) {
        return -1;
      }

      inet_prefix dst;
      get_prefix(&dst, *argv, req.r.rtm_family);

      if (req.r.rtm_family == AF_UNSPEC) {
        req.r.rtm_family = dst.family;
      }
      req.r.rtm_dst_len = dst.bitlen;
      dst_ok = 1;
      if (dst.bytelen) {
        addattr_l(&req.n, sizeof(req), RTA_DST, &dst.data, dst.bytelen);
      }
    }
    _NEXT_ARG;
  }

  if (d) {
    int idx = ll_name_to_index(d);
    if (!idx) {
      return -1;
    }
    addattr32(&req.n, sizeof(req), RTA_OIF, idx);
  }

  if (req.r.rtm_family == AF_UNSPEC) {
    req.r.rtm_family = AF_INET;
  }

  if (!table_ok && (
          req.r.rtm_type == RTN_LOCAL ||
          req.r.rtm_type == RTN_BROADCAST ||
          req.r.rtm_type == RTN_NAT ||
          req.r.rtm_type == RTN_ANYCAST)) {
    req.r.rtm_table = RT_TABLE_LOCAL;
  }

  if (!scope_ok) {
    if (req.r.rtm_family == AF_INET6 ||
        req.r.rtm_family == AF_MPLS) {
      req.r.rtm_scope = RT_SCOPE_UNIVERSE;
    } else if (req.r.rtm_type == RTN_LOCAL ||
               req.r.rtm_type == RTN_NAT) {
      req.r.rtm_scope = RT_SCOPE_HOST;
    } else if (req.r.rtm_type == RTN_BROADCAST ||
               req.r.rtm_type == RTN_MULTICAST ||
               req.r.rtm_type == RTN_ANYCAST) {
      req.r.rtm_scope = RT_SCOPE_LINK;
    } else if (req.r.rtm_type == RTN_UNICAST ||
               req.r.rtm_type == RTN_UNSPEC) {
      if (msg_type == RTM_DELROUTE) {
        req.r.rtm_scope = RT_SCOPE_NOWHERE;
      } else if (!gw_ok && !nhs_ok) {
        req.r.rtm_scope = RT_SCOPE_LINK;
      }
    }
  }

  if (!type_ok && req.r.rtm_family == AF_MPLS) {
    req.r.rtm_type = RTN_UNICAST;
  }

  return send_rtnl_message(&req.n);
}


int
iproute_add(int argc, char **argv) {
  return iproute_modify(RTM_NEWROUTE, NLM_F_CREATE|NLM_F_EXCL, argc, argv);
}
