#include "base.h"
#include "mount.h"
#include <string.h>
#include <sys/mount.h>


struct mount_atts {
  const char *src;
  const char *target;
  const char *fstype;
  unsigned long flags;
};


static struct mount_atts mfiles[] = {
  {"proc",    "proc",     "proc",     0},
  {"sysfs",   "sys",      "sysfs",    0},
  {"none",    "tmp",      "tmpfs",    0},
  {"udev",    "dev",      "devtmpfs", 0},
  {"devpts",  "dev/pts",  "devpts",   0},
  {"shm",     "dev/shm",  "tmpfs",    0},
  {"tmpfs",   "run",      "tmpfs",    0},

  /* {"conf/hosts",       "etc/hosts",       "none", MS_BIND}, */
  /* {"conf/hostname",    "etc/hostname",    "none", MS_BIND}, */
  /* {"conf/resolv.conf", "etc/resolv.conf", "none", MS_BIND}, */
};


int
mount_fs(const char *path) {
  size_t n = sizeof(mfiles) / sizeof(struct mount_atts);

  char p[64];
  size_t len = strlen(path);
  strcpy(p, path);

  for (size_t i = 0; i < n; ++i) {
    struct mount_atts *f = mfiles + i;
    int r;

    strcpy(p + len, mfiles[i].target);
    if ((r = mount(f->src, p, f->fstype, f->flags, NULL)) != 0) {
      lperror(p);
      return r;
    }
  }

  return 0;
}


int
umount_fs(const char *path) {
  size_t n = sizeof(mfiles) / sizeof(struct mount_atts);

  char p[64];
  size_t len = strlen(path);
  strcpy(p, path);

  for (int i = n - 1; i >= 0; --i) {
    int r;
    strcpy(p + len, mfiles[i].target);
    if ((r = umount(p)) != 0) {
      lperror(p);
    }
  }

  return 0;
}
