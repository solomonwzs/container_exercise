#include "network.h"
#include <linux/veth.h>
#include <stdio.h>
#include <stdlib.h>
#include <net/if.h>


struct iplink_req {
  struct nlmsghdr   n;
  struct ifinfomsg  i;
  char              buf[1024];
};

static struct rtnl_handle rth = {.fd = -1};


void
foo() {
  printf("hello world\n");
}

int
net_create_veth(const char *dev, const char *nsdev, unsigned pid) {
  if (rtnl_open(&rth, 0) < 0) {
    fprintf(stderr, "cannot open netlink\n");
    return 1;
  }

  struct iplink_req req;
  memset(&req, 0, sizeof(req));
  req.n.nlmsg_len = NLMSG_LENGTH(sizeof(struct ifinfomsg));
  req.n.nlmsg_flags = NLM_F_REQUEST|NLM_F_CREATE|NLM_F_EXCL;
  req.n.nlmsg_type = RTM_NEWLINK;
  req.i.ifi_family = 0;

  if (dev) {
    addattr_l(&req.n, sizeof(req), IFLA_IFNAME, dev, strlen(dev) + 1);
  }

  // add link info for the new interface
  struct rtattr *linkinfo = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_LINKINFO, NULL, 0);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_KIND, "veth", strlen("veth"));

  struct rtattr *data = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_DATA, NULL, 0);

  struct rtattr *peerdata = NLMSG_TAIL(&req.n);
  addattr_l (&req.n, sizeof(req), VETH_INFO_PEER, NULL, 0);
  req.n.nlmsg_len += sizeof(struct ifinfomsg);

  // place the link in the child namespace
  addattr_l(&req.n, sizeof(req), IFLA_NET_NS_PID, &pid, sizeof(pid));

  if (nsdev) {
    addattr_l(&req.n, sizeof(req), IFLA_IFNAME, nsdev, strlen(nsdev));
  }

  peerdata->rta_len = (void *)NLMSG_TAIL(&req.n) - (void *)peerdata;
  data->rta_len = (void *)NLMSG_TAIL(&req.n) - (void *)data;
  linkinfo->rta_len = (void *)NLMSG_TAIL(&req.n) - (void *)linkinfo;

  // send message
  if (rtnl_talk(&rth, &req.n, NULL) < 0) {
    return 1;
  }
  rtnl_close(&rth);

  return 0;
}


int
net_create_ipvlan(const char *host_dev, const char *dev, uint16_t type,
                  unsigned pid) {
  if (rtnl_open(&rth, 0) < 0) {
    fprintf(stderr, "cannot open netlink\n");
    return 1;
  }

  struct iplink_req req;
  memset(&req, 0, sizeof(req));
  req.n.nlmsg_len = NLMSG_LENGTH(sizeof(struct ifinfomsg));
  req.n.nlmsg_flags = NLM_F_REQUEST|NLM_F_CREATE|NLM_F_EXCL;
  req.n.nlmsg_type = RTM_NEWLINK;
  req.i.ifi_family = 0;

  // find host ifindex
  int host_ifindex = if_nametoindex(host_dev);
  if (host_ifindex <= 0) {
    fprintf(stderr, "Error: cannot find network device %s\n", host_dev);
    exit(1);
  }

  // add host
  addattr_l(&req.n, sizeof(req), IFLA_LINK, &host_ifindex,
            sizeof(host_ifindex));

  // add new interface name
  addattr_l(&req.n, sizeof(req), IFLA_IFNAME, dev, strlen(dev) + 1);

  // place the link in the child namespace
  addattr_l(&req.n, sizeof(req), IFLA_NET_NS_PID, &pid, sizeof(pid));

  // add link info for the new interface
  struct rtattr *linkinfo = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_LINKINFO, NULL, 0);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_KIND, "ipvlan", strlen("ipvlan"));

  struct rtattr *data = NLMSG_TAIL(&req.n);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_DATA, NULL, 0);
  addattr_l(&req.n, sizeof(req), IFLA_INFO_KIND, &type, sizeof(type));

  data->rta_len = (void *)NLMSG_TAIL(&req.n) - (void *)data;
  linkinfo->rta_len = (void *)NLMSG_TAIL(&req.n) - (void *)linkinfo;

  // send message
  if (rtnl_talk(&rth, &req.n, NULL) < 0) {
    return 1;
  }
  rtnl_close(&rth);

  return 0;
}
