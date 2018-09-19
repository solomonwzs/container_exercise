#define _GNU_SOURCE

#include "base.h"
#include "capability.h"
#include "container.h"
#include "mount.h"
#include "id_map.h"
#include <sched.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <sys/wait.h>

#define STACK_SIZE (1024 * 1024)
#define HOSTNAME "container"

#define get_ns(_pid_, _ns_) do { \
  char __filename[64]; \
  char __buf[64]; \
  sprintf(__filename, "/proc/%d/ns/%s", _pid_, _ns_); \
  readlink(__filename, __buf, 64); \
  ldebug("%d: %s\n", _pid_, __buf); \
} while (0)


struct container_arg {
  const char *basesystem_path;
  int pipefd[2];
};


static int
run(void *arg) {
  ldebug("Container start.\n");

  struct container_arg *carg = (struct container_arg *)arg;
  char ch;
  close(carg->pipefd[1]);
  list_caps();
  read(carg->pipefd[0], &ch, 1);

  if (sethostname(HOSTNAME, strlen(HOSTNAME)) == -1) {
    lperror("sethostname");
    // return 1;
  }

  const char *path = carg->basesystem_path;
  umount_fs(path);
  if (mount_fs(path) != 0) {
    // return 1;
  }

  if (chdir(path) != 0 || chroot("./") != 0) {
    lperror("chdir/chroot");
  }

  char *cmd[] = {
    "/bin/bash",
    NULL
  };
  execv(cmd[0], cmd);
  lperror("execv");

  return 1;
}


int
container_run(const char *path) {
  ldebug("Start.\n");

  struct container_arg carg;
  carg.basesystem_path = path;
  if (pipe(carg.pipefd) != 0) {
    lperror("pipe");
    return 1;
  }

  u_int8_t stack[STACK_SIZE];
  pid_t self = getpid();
  pid_t container_pid = clone(run, stack + STACK_SIZE,
                              CLONE_NEWUSER   // User namespaces
                              | CLONE_NEWNS   // Mount namespaces
                              | CLONE_NEWIPC  // IPC namespaces
                              | CLONE_NEWPID  // PID namespaces
                              | CLONE_NEWUTS  // UTS namespaces
                              | SIGCHLD,
                              &carg);

  if (container_pid == -1) {
    lperror("new container");
    return 1;
  }
  ldebug("Container PID: %d\n", container_pid);

  set_uid_map(container_pid, 0, getuid(), 1);
  set_gid_map(container_pid, 0, getgid(), 1);
  close(carg.pipefd[1]);

  get_ns(self, "pid");
  get_ns(container_pid, "pid");

  waitpid(container_pid, NULL, 0);

  ldebug("End.\n");
  return 0;
}
