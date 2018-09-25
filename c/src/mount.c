#include "base.h"
#include "mount.h"
#include <string.h>
#include <sys/mount.h>


static const struct mount_attrs mfiles[] = {
  {"proc",    "proc",     "proc",     0},
  {"sysfs",   "sys",      "sysfs",    0},
  {"none",    "tmp",      "tmpfs",    0},
  {"udev",    "dev",      "devtmpfs", 0},
  {"devpts",  "dev/pts",  "devpts",   0},
  {"shm",     "dev/shm",  "tmpfs",    0},
  {"tmpfs",   "run",      "tmpfs",    0},

  {"/etc/hosts",       "etc/hosts",       "none", MS_BIND},
  // {"/etc/hostname",    "etc/hostname",    "none", MS_BIND},
  {"/etc/resolv.conf", "etc/resolv.conf", "none", MS_BIND},
};


int
mount_fs(const char *path) {
  size_t n = sizeof(mfiles) / sizeof(struct mount_attrs);

  char p[64];
  size_t len = strlen(path);
  strcpy(p, path);

  for (int i = 0; i < n; ++i) {
    const struct mount_attrs *f = mfiles + i;
    int r;

    strcpy(p + len, mfiles[i].target);
    if ((r = mount(f->src, p, f->fstype, f->flags, NULL)) != 0) {
      lperror(p);
      // return r;
    } else {
      ldebug("%s: ok\n", p);
    }
  }

  return 0;
}


int
umount_fs(const char *path) {
  size_t n = sizeof(mfiles) / sizeof(struct mount_attrs);

  char p[64];
  size_t len = strlen(path);
  strcpy(p, path);

  for (int i = n - 1; i >= 0; --i) {
    int r;
    strcpy(p + len, mfiles[i].target);
    if ((r = umount(p)) != 0) {
      lperror(p);
    } else {
      ldebug("%s: ok\n", p);
    }
  }

  return 0;
}
