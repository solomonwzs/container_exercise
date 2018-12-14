#ifndef NETWORK_H
#define NETWORK_H

extern void
foo();

extern int
net_create_veth(const char *dev, const char *nsdev, unsigned pid);

extern int
net_create_ipvlan(const char *host_dev, const char *dev, int type,
                  unsigned pid);

#endif
