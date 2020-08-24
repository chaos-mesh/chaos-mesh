#include <signal.h>
#include <unistd.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>
#include <sys/prctl.h>

int main(int argc, char* argv[]) {
  if (prctl(PR_SET_CHILD_SUBREAPER, 1) < 0) {
    fprintf(stderr, "fail to SET_CHILD_SUBREAPER %s", strerror(errno));
    return -1;
  }

  int ret = execvp(argv[1], &argv[1]);
  if (ret == -1) {
    fprintf(stderr, "%s", strerror(errno));
    return -1;
  }
  return 0;
}
