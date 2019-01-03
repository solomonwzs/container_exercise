#ifndef NETWORK_H
#define NETWORK_H

#include <stdint.h>
#include <libnetlink.h>

#ifndef IFLA_IPVLAN_MAX
#define IPVLAN_MODE_L2 0
#endif

#define _NEXT_ARG do { argc--; argv++; } while (0)

extern int
send_rtnl_message(struct nlmsghdr *n);

extern int
iplink_create_veth(const char *dev, const char *nsdev, unsigned pid);

extern int
iplink_create_vlan(const char *host_dev, const char *dev, unsigned pid,
                   uint16_t vid);
extern int
iplink_create_ipvlan(const char *host_dev, const char *dev, unsigned pid,
                     uint16_t type);

extern int
iplink_create_macvlan(const char *host_dev, const char *dev, unsigned pid,
                      uint32_t type);

extern int
iplink_create_bridge(const char *dev);

extern int
iplink_set_master(const char *dev, const char *masterdev);

extern int
iplink_delete_dev(const char *dev);

extern int
iplink_rename(const char *dev, const char *newdev);

extern int
iplink_chflags(const char *dev, uint32_t flags, uint32_t mask);

extern int
iproute_add(int argc, char **argv);

extern int
ipaddr_add(int argc, char **argv);

#endif
