#ifndef ID_MAP_H
#define ID_MAP_H

#include <sys/types.h>

extern int set_uid_map(pid_t pid, int inside_id, int outside_id, int len);

extern int set_gid_map(pid_t pid, int inside_id, int outside_id, int len);

#endif
