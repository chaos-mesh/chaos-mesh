#include <signal.h>
#include <unistd.h>
#include <stdio.h>
#include <string.h>
#include <errno.h>

int main(int argc, char* argv[]) {
    raise(SIGSTOP);
    int ret = execvp(argv[1], &argv[1]);
    if (ret == -1) {
		fprintf(stderr, "%s", strerror(errno));
		return -1;
	}
	return 0;
}