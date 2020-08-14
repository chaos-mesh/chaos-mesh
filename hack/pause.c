#include <signal.h>
#include <unistd.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>

static void handler() {}

int main(int argc, char* argv[]) {
  signal(SIGCONT, handler);
  pause();

  int ret = execvp(argv[1], &argv[1]);
  if (ret == -1) {
    fprintf(stderr, "%s", strerror(errno));
    return -1;
  }
  return 0;
}
