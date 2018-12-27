#ifndef NETWORK_H
#define NETWORK_H

#include <stdint.h>
#include <libnetlink.h>

#ifndef IFLA_IPVLAN_MAX
#define IPVLAN_MODE_L2 0
#endif

extern int
iplink_create_veth(const char *dev, const char *nsdev, unsigned pid);

extern int
iplink_create_ipvlan(const char *host_dev, const char *dev, uint16_t type,
                  unsigned pid);

extern int
iplink_rename(const char *dev, const char *newdev);

extern int
iplink_chflags(const char *dev, uint32_t flags, uint32_t mask);

extern int
iplink_create_bridge(const char *dev);

#endif
