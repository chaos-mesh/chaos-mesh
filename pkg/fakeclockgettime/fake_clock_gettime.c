#include <time.h>
#include <inttypes.h>

int64_t TV_SEC_DELTA = 0;
int64_t TV_NSEC_DELTA = 0;

long syscall(long number, ...);

int clock_gettime(clockid_t clk_id, struct timespec *tp) {
    int ret;
    asm volatile
        (
            "syscall"
            : "=a" (ret)
            : "0"(228), "D"(clk_id), "S"(tp)
            : "rcx", "r11", "memory"
        );

    if(clk_id == CLOCK_REALTIME) {
        tp->tv_sec += TV_SEC_DELTA;

        // TODO: avoid overflow here!!!
        tp->tv_nsec += TV_NSEC_DELTA;
    }

    return ret;
}