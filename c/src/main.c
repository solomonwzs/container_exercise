#define _GNU_SOURCE

#include "base.h"
#include "capability.h"
#include "container.h"
#include "mount.h"
#include "netns.h"
#include <sched.h>
#include <signal.h>
#include <stdlib.h>
#include <sys/wait.h>

#define STACK_SIZE (1024 * 1024)


int
run(void *arg) {
  return 0;
}


int
test_ns() {
  netns_switch("ns1");
  system("ip addr");
  system("ping -c 1 127.0.0.1");

  return 0;
}


int
main(int argc, char **argv) {
  // return test_ns();

  return container_run(argc, argv);

  // char stack[STACK_SIZE];
  // pid_t pid = clone(run, stack + STACK_SIZE,
  //                   CLONE_NEWUSER | SIGCHLD, NULL);
  // if (pid == -1) {
  //   lperror("clone");
  //   return 1;
  // } else {
  //   waitpid(pid, NULL, 0);
  // }
}
