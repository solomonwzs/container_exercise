#include "base.h"
#include "id_map.h"


static inline int
set_map(const char *filename, int inside_id, int outside_id, int len) {
  FILE *fd = fopen(filename, "w");
  if (fd == NULL) {
    lperror("open");
    return 1;
  }

  int r = fprintf(fd, "%d %d %d", inside_id, outside_id, len) < 0 ? 1 : 0;
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
