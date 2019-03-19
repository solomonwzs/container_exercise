#ifndef CONT_PROTO_H
#define CONT_PROTO_H

#include <stdint.h>

enum cont_opcode {
  CONT_INIT = 1,
};

struct cont_in_header {
  uint32_t opcode;
  uint64_t unique;
  uint32_t len;
};

#endif
