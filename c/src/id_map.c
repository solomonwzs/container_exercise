#include "base.h"
#include "id_map.h"
#include "pathnames.h"


static inline int
set_map(const char *filename, int inside_id, int outside_id, int len) {
  FILE *fd = fopen(filename, "w");
  if (fd == NULL) {
    lperror(filename);
    return -1;
  }

  int r = fprintf(fd, "%d %d %d", inside_id, outside_id, len) < 0 ? -1 : 0;
  fclose(fd);
  return r;
}


int
set_uid_map(pid_t pid, int inside_id, int outside_id, int len) {
  char filename[64];
  sprintf(filename, "/proc/%d/uid_map", pid);
  return set_map(filename, inside_id, outside_id, len);
}


int
set_gid_map(pid_t pid, int inside_id, int outside_id, int len) {
  char filename[64];
  sprintf(filename, "/proc/%d/gid_map", pid);
  return set_map(filename, inside_id, outside_id, len);
}


int setgroups_ctrl(int action) {
  FILE *fd = fopen(_PATH_PROC_SETGROUPS, "w");
  if (fd == NULL) {
    lperror(_PATH_PROC_SETGROUPS);
    return -1;
  }
  const char *cmd = action == SETGROUPS_DENY ? "deny" : "allow";
  int r = fprintf(fd, cmd) < 0 ? -1 : 0;
  fclose(fd);
  return r;
}
