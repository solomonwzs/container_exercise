#include "network.h"
#include "uapi/linux/sockios.h"

#include <linux/veth.h>
#include <net/if.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stropts.h>
#include <unistd.h>

#define SET_RTA_LEN(_rta, _hdr) do { \
  (_rta)->rta_len = (void *)NLMSG_TAIL(_hdr) - (void *)(_rta); \
} while (0)


struct iplink_req {
  struct nlmsghdr   n;
  struct ifinfomsg  i;
  char              buf[1024];
};


static inline void
init_iplink_req(struct iplink_req *req, uint16_t type, uint16_t flags) {
  memset(req, 0, sizeof(struct iplink_req));
  req->n.nlmsg_len = NLMSG_LENGTH(sizeof(struct ifinfomsg));
  req->n.nlmsg_type = type;
  req->n.nlmsg_flags = NLM_F_REQUEST|flags;
  req->i.ifi_family = 0;
}


static inline int
get_ctl_fd() {
  int fd;

  if ((fd = socket(PF_INET, SOCK_DGRAM, 0)) >= 0) {
    return fd;
  }
  if ((fd = socket(PF_PACKET, SOCK_DGRAM, 0)) >= 0) {
    return fd;
  }
  if ((fd = socket(PF_INET6, SOCK_DGRAM, 0)) >= 0) {
    return fd;
  }

  return -1;
}


static inline int
send_message(struct nlmsghdr *n) {
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


static int
iplink_create_xvlan(const char *host_dev, const char *dev, unsigned pid,
                    const char *type, void *attri) {
  int attr;
  int len;
  if (strcmp(type, "vlan") == 0) {
    attr = IFLA_VLAN_ID;
    len = 2;
  } else if (strcmp(type, "ipvlan") == 0) {
    attr = IFLA_INFO_KIND;
    len = 2;
  } else if (strcmp(type, "macvlan") == 0) {
    attr = IFLA_INFO_KIND;
    len = 4;
  } else {
    return -1;
  }

  struct iplink_req req;
  init_iplink_req(&req, RTM_NEWLINK, NLM_F_CREATE|NLM_F_EXCL);

  // find host ifindex
  int host_ifindex = if_nametoindex(host_dev);
  if (host_ifindex <= 0) {
    return -1;
  }
  addattr_l(&req.n, sizeof(req), IFLA_LINK, &host_ifindex,
            sizeof(host_ifindex));

  // add new interface name
  addattr_l(&req.n, sizeof(req), IFLA_IFNAME, dev, strlen(dev) + 1);
  // place the link in the child namespace
  addattr_l(&req.n, sizeof(req), IFLA_NET_NS_PID, &pid, sizeof(pid));

  struct rtattr *linkinfo = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_LINKINFO, NULL, 0);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_KIND, type, strlen(type));

  struct rtattr *data = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_DATA, NULL, 0);
  addattr_l(&req.n, sizeof(req), attr, attri, len);

  SET_RTA_LEN(data, &req.n);
  SET_RTA_LEN(linkinfo, &req.n);

  return send_message(&req.n);
}


int
iplink_create_vlan(const char *host_dev, const char *dev, unsigned pid,
                   uint16_t vid) {
  return iplink_create_xvlan(host_dev, dev, pid, "vlan", &vid);
}


int
iplink_create_ipvlan(const char *host_dev, const char *dev, unsigned pid,
                     uint16_t type) {
  return iplink_create_xvlan(host_dev, dev, pid, "ipvlan", &type);
}


int
iplink_create_macvlan(const char *host_dev, const char *dev, unsigned pid,
                      uint32_t type) {
  return iplink_create_xvlan(host_dev, dev, pid, "macvlan", &type);
}


int
iplink_delete_dev(const char *dev) {
  int ifindex = if_nametoindex(dev);
  if (ifindex <= 0) {
    return -1;
  }

  struct iplink_req req;
  init_iplink_req(&req, RTM_DELLINK, 0);
  req.i.ifi_index = ifindex;

  return send_message(&req.n);
}


int
iplink_create_bridge(const char *dev) {
  struct iplink_req req;
  init_iplink_req(&req, RTM_NEWLINK, NLM_F_CREATE|NLM_F_EXCL);

  addattr_l(&req.n, sizeof(req), IFLA_IFNAME, dev, strlen(dev) + 1);

  struct rtattr *linkinfo = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_LINKINFO, NULL, 0);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_KIND, "bridge", strlen("bridge"));
  SET_RTA_LEN(linkinfo, &req.n);

  return send_message(&req.n);
}


int
iplink_set_master(const char *dev, const char *masterdev) {
  int ifindex = if_nametoindex(dev);
  if (ifindex <= 0) {
    return -1;
  }

  struct iplink_req req;
  init_iplink_req(&req, RTM_NEWLINK, 0);
  req.i.ifi_index = ifindex;

  int mifindex = if_nametoindex(masterdev);
  if (mifindex <= 0) {
    return -1;
  }
  addattr_l(&req.n, sizeof(req), IFLA_MASTER, &mifindex, sizeof(mifindex));

  return send_message(&req.n);
}


int
iplink_create_veth(const char *dev, const char *nsdev, unsigned pid) {
  struct iplink_req req;
  init_iplink_req(&req, RTM_NEWLINK, NLM_F_CREATE|NLM_F_EXCL);

  addattr_l(&req.n, sizeof(req), IFLA_IFNAME, dev, strlen(dev) + 1);

  // add link info for the new interface
  struct rtattr *linkinfo = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_LINKINFO, NULL, 0);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_KIND, "veth", strlen("veth"));

  struct rtattr *data = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_DATA, NULL, 0);

  struct rtattr *peerdata = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), VETH_INFO_PEER, NULL, 0);
  req.n.nlmsg_len += sizeof(struct ifinfomsg);

  // place the link in the child namespace
  addattr_l(&req.n, sizeof(req), IFLA_NET_NS_PID, &pid, sizeof(pid));
  addattr_l(&req.n, sizeof(req), IFLA_IFNAME, nsdev, strlen(nsdev));

  SET_RTA_LEN(peerdata, &req.n);
  SET_RTA_LEN(data, &req.n);
  SET_RTA_LEN(linkinfo, &req.n);

  return send_message(&req.n);
}


int
iplink_rename(const char *dev, const char *newdev) {
  struct ifreq ifr;
  strncpy(ifr.ifr_ifrn.ifrn_name, dev, IF_NAMESIZE);
  strncpy(ifr.ifr_ifru.ifru_newname, newdev, IF_NAMESIZE);

  int fd = get_ctl_fd();
  if (fd < 0) {
    return -1;
  }

  if (ioctl(fd, SIOCSIFNAME, &ifr) != 0) {
    close(fd);
    return -1;
  }

  close(fd);
  return 0;
}


int
iplink_chflags(const char *dev, uint32_t flags, uint32_t mask) {
  int ifindex = if_nametoindex(dev);
  if (ifindex <= 0) {
    return -1;
  }

  struct iplink_req req;
  init_iplink_req(&req, RTM_NEWLINK, 0);
  req.i.ifi_change = mask;
  req.i.ifi_flags = flags;
  req.i.ifi_index = ifindex;

  return send_message(&req.n);
}
