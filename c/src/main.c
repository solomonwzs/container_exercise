#define _GNU_SOURCE

#include "base.h"
#include <sched.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mount.h>
#include <sys/wait.h>

#define STACK_SIZE (1024 * 1024)
#define HOSTNAME "container"

static int
container_run(void *arg) {
  ldebug("Container start.\n");

  if (sethostname(HOSTNAME, strlen(HOSTNAME)) == -1) {
    lperror("sethostname");
    return 1;
  }
  if (mount("proc", "/proc", "proc", 0, NULL) != 0) {
    lperror("mount");
    return 1;
  }

  char *cmd[] = {
    "/bin/bash",
    NULL
  };
  execv(cmd[0], cmd);
  lperror("execv");

  return 0;
}


int
main(int argc, char **args) {
  ldebug("Start.\n");

  u_int8_t stack[STACK_SIZE];
  pid_t container_pid = clone(container_run, stack + STACK_SIZE,
                              SIGCHLD
                              | CLONE_NEWNS   // Mount namespaces
                              | CLONE_NEWIPC  // IPC namespaces
                              | CLONE_NEWPID  // PID namespaces
                              | CLONE_NEWUTS, // UTS namespaces
                              NULL);

  if (container_pid == -1) {
    lperror("new container");
  } else {
    waitpid(container_pid, NULL, 0);
  }

  ldebug("End.\n");
  return 0;
}
