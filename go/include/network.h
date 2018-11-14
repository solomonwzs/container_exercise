#ifndef NETWORK_H
#define NETWORK_H

extern void foo();

extern int net_create_veth(const char *dev, const char *nsdev, unsigned pid);

#endif
