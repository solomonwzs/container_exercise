#ifndef _MY_ID_MAP_H
#define _MY_ID_MAP_H

#include <sys/types.h>

#define SETGROUPS_DENY  0x00
#define SETGROUPS_ALLOW 0x01

extern int set_uid_map(pid_t pid, int inside_id, int outside_id, int len);

extern int set_gid_map(pid_t pid, int inside_id, int outside_id, int len);

extern int setgroups_ctrl(int action);

#endif
