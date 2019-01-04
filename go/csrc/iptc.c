#include <errno.h>
#include <libiptc/libiptc.h>
#include <string.h>
#include <stdio.h>


struct rule_entry {
  struct ipt_entry          e;
  struct xt_standard_target t;
};


int
iptc_add_rule(const char *table, const char *chain,
              unsigned src, unsigned smsk, int sinv,
              unsigned dst, unsigned dmsk, int dinv,
              const char *target) {
  struct rule_entry entry;
  struct xtc_handle *h;
  int ret = -1;

  if (!(h = iptc_init(table))) {
    goto release;
  }
  memset(&entry, 0, sizeof(entry));

  entry.t.target.u.user.target_size = XT_ALIGN(
      sizeof(struct xt_standard_target));
  strncpy(entry.t.target.u.user.name, target,
          sizeof(entry.t.target.u.user.name));

  entry.e.target_offset = sizeof(struct ipt_entry);
  entry.e.next_offset = entry.e.target_offset +
      entry.t.target.u.user.target_size;

  if (src) {
    entry.e.ip.src.s_addr = src;
    entry.e.ip.smsk.s_addr = smsk;
    if (sinv) {
      entry.e.ip.invflags |= IPT_INV_SRCIP;
    }
  }

  if (dst) {
    entry.e.ip.dst.s_addr  = dst;
    entry.e.ip.dmsk.s_addr = dmsk;
    if (dinv) {
      entry.e.ip.invflags |= IPT_INV_DSTIP;
    }
  }

  if (!iptc_append_entry(chain, (struct ipt_entry *)&entry, h)) {
    fprintf(stderr, "%s\n", iptc_strerror(errno));
    goto release;
  }

  if (!iptc_commit(h)) {
    fprintf(stderr, "%s\n", iptc_strerror(errno));
    goto release;
  }

  ret = 0;

release:
  if (h) { iptc_free(h); }

  return ret;
}
