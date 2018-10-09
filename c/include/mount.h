#ifndef _MY_MOUNT_H
#define _MY_MOUNT_H

struct mount_attrs {
  const char *src;
  const char *target;
  const char *fstype;
  unsigned long flags;
};

extern int mount_fs(const char *path);

extern int umount_fs(const char *path);

#endif
