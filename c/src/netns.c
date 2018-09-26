#define _GNU_SOURCE

#include "base.h"
#include "netns.h"
#include "pathnames.h"
#include <fcntl.h>
#include <sys/mount.h>
#include <sched.h>
#include <stdio.h>
#include <sys/statvfs.h>


int
netns_switch(const char *ns_name) {
  char net_path[_PATH_MAX_LEN];
  int netns;

  snprintf(net_path, sizeof(net_path), "%s/%s", _PATH_NETNS_RUN_DIR,
           ns_name);
  if ((netns = open(net_path, O_RDONLY | O_CLOEXEC)) < 0) {
    lperror("open");
    return -1;
  }

  if (setns(netns, CLONE_NEWNET) < 0) {
    lperror("setns");
    close(netns);
    return -1;
  }
  close(netns);

  if (unshare(CLONE_NEWNS) < 0) {
    lperror("unshare");
    return -1;
  }

  // if (mount("", "/", "none", MS_SLAVE | MS_REC, NULL) < 0) {
  //   lperror("mount /");
  //   return -1;
  // }

  // int mount_flags = 0;
  // if (umount2("/sys", MNT_DETACH) < 0) {
  //   lperror("umount2 /sys");

  //   struct statvfs fs_stat;
  //   if (statvfs("/sys", &fs_stat) == 0) {
  //     if (fs_stat.f_flag & ST_RDONLY) {
  //       mount_flags = MS_RDONLY;
  //     }
  //   }
  // }

  // if (mount(ns_name, "/sys", "sysfs", mount_flags, NULL) < 0) {
  //   lperror("mount /sys");
  //   return -1;
  // }

  return 0;
}
