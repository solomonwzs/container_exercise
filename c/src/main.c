#define _GNU_SOURCE

#include "base.h"
#include "container.h"
#include "capability.h"
#include <sched.h>
#include <signal.h>
#include <sys/wait.h>

#define STACK_SIZE (1024 * 1024)


int
run(void *arg) {
  return 0;
}


int
main(int argc, char **args) {
  if (argc < 2) {
    ldebug("miss arguments.\n");
    return 1;
  }
  const char *path = args[1];
  container_run(path);

  // char stack[STACK_SIZE];
  // pid_t pid = clone(run, stack + STACK_SIZE,
  //                   CLONE_NEWUSER | SIGCHLD, NULL);
  // if (pid == -1) {
  //   lperror("clone");
  //   return 1;
  // } else {
  //   waitpid(pid, NULL, 0);
  // }

  return 0;
}
