#ifndef CONT_PROTO_H
#define CONT_PROTO_H

#include <stdint.h>

enum cont_opcode {
  CONT_INIT = 1,
};

struct cont_in_header {
  uint64_t unique;
  uint32_t opcode;
  uint32_t len;
};

struct cont_init_in {
  uint32_t pid;
  uint32_t padding;
};

struct cont_init_out {
  uint32_t pid;
  uint32_t padding;
};

#endif
