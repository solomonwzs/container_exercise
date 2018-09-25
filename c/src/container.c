#define _GNU_SOURCE

#include "base.h"
#include "capability.h"
#include "container.h"
#include "id_map.h"
#include "mount.h"
#include <sched.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mount.h>
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


static struct env_attrs {
  const char *name;
  const char *value;
  int overwrite;
} container_env[] = {
  {"PATH", "/usr/bin:/usr/local/bin:/usr/local/sbin:/bin:/sbin", 1},
};


struct container_arg {
  int argc;
  char **argv;
  int pipefd[2];
};


static inline void
child_wait(struct container_arg *arg) {
  char ch;
  read(arg->pipefd[0], &ch, 1);
}


static inline void
child_awake(struct container_arg *arg) {
  char ch;
  write(arg->pipefd[1], &ch, 1);
}


static int
run(void *arg) {
  ldebug("Container start.\n");

  struct container_arg *carg = (struct container_arg *)arg;
  close(carg->pipefd[1]);
  child_wait(carg);

  int opt;
  int len = 0;
  char buf[128];
  char path[128] = {0};
  while ((opt = getopt(carg->argc, carg->argv, "+n:m:p:e:")) != -1) {
    switch (opt) {
      case 'n':
        if (sethostname(optarg, strlen(optarg)) == -1) {
          lperror("hostname");
          return 1;
        }
        break;
      case 'p':
        strcpy(path, optarg);
        len = strlen(path);
        break;
      case 'm':
        if (len == 0) {
          ldebug("please set -p first\n");
          return 1;
        }

        for (int i = 0; i < strlen(optarg); ++i) {
          if (optarg[i] == ':') {
            strncpy(buf, optarg, i);
            strcpy(path + len, optarg + i + 1);

            const char *src = buf;
            const char *target = path;

            FILE *fd = fopen(target, "a");
            if (fd == NULL) {
              lperror(target);
              return 1;
            } else {
              fclose(fd);
            }

            if (mount(src, target, "none", MS_BIND, NULL) != 0) {
              lperror("mount");
            }
            break;
          }
        }
        break;
      case 'e':
        break;
      default:
        return 1;
    }
  }
  path[len] = '\0';

  // cap_t caps = cap_from_text("all= cap_sys_admin-e cap_net_raw+ep");
  // cap_t caps = cap_from_text("all+ep cap_net_raw-ep");
  // cap_set_proc(caps);
  // cap_free(caps);
  list_caps;

  if (mount_fs(path) != 0) {
    return 1;
  }
  if (chdir(path) != 0 || chroot("./") != 0) {
    lperror("chdir/chroot");
    return 1;
  }

  size_t n = sizeof(container_env) / sizeof(struct env_attrs);
  for (int i = 0; i < n; ++i) {
    struct env_attrs *e = container_env + i;
    setenv(e->name, e->value, e->overwrite);
  }

  system("/bin/ping -c 1 baidu.com");

  char *cmd[] = {
    "/bin/bash",
    NULL
  };
  execvp(cmd[0], cmd);
  lperror("execv");

  return 1;
}


int
container_run(int argc, char **argv) {
  ldebug("Start.\n");

  struct container_arg carg = {
    .argc = argc,
    .argv = argv,
  };
  if (pipe(carg.pipefd) != 0) {
    lperror("pipe");
    return 1;
  }

  u_int8_t stack[STACK_SIZE];
  pid_t self = getpid();
  pid_t container_pid = clone(run, stack + STACK_SIZE,
                              CLONE_NEWNS     // Mount namespaces
                              | CLONE_NEWNET  // Network namespaces
                              | CLONE_NEWUSER // User namespaces
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

  close(carg.pipefd[0]);
  child_awake(&carg);

  get_ns(self, "pid");
  get_ns(container_pid, "pid");

  waitpid(container_pid, NULL, 0);
  close(carg.pipefd[1]);

  ldebug("End.\n");
  return 0;
}
