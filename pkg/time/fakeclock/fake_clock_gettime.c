/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
#include <time.h>
#include <inttypes.h>
#include <syscall.h>

extern int64_t TV_SEC_DELTA;
extern int64_t TV_NSEC_DELTA;
extern uint64_t CLOCK_IDS_MASK;

#if defined(__amd64__)
inline int real_clock_gettime(clockid_t clk_id, struct timespec *tp) {
    int ret;
    asm volatile
        (
            "syscall"
            : "=a" (ret)
            : "0"(__NR_clock_gettime), "D"(clk_id), "S"(tp)
            : "rcx", "r11", "memory"
        );

    return ret;
}
#elif defined(__aarch64__)
inline int real_clock_gettime(clockid_t clk_id, struct timespec *tp) {
    register clockid_t x0 __asm__ ("x0") = clk_id;
    register struct timespec *x1 __asm__ ("x1") = tp;
    register uint64_t w8 __asm__ ("w8") = __NR_clock_gettime; /* syscall number */
    __asm__ __volatile__ (
        "svc 0;"
        : "+r" (x0)
        : "r" (x0), "r" (x1), "r" (w8)
        : "memory"
    );

    return x0;
}
#endif

int fake_clock_gettime(clockid_t clk_id, struct timespec *tp) {
    int ret = real_clock_gettime(clk_id, tp);

    int64_t sec_delta = TV_SEC_DELTA;
    int64_t nsec_delta = TV_NSEC_DELTA;
    uint64_t clock_ids_mask = CLOCK_IDS_MASK;

    int64_t billion = 1000000000;

    uint64_t clk_id_mask = 1 << clk_id;
    if((clk_id_mask & clock_ids_mask) != 0) {
        while (nsec_delta + tp->tv_nsec > billion) {
            sec_delta += 1;
            nsec_delta -= billion;
        }

        while (nsec_delta + tp->tv_nsec < 0) {
            sec_delta -= 1;
            nsec_delta += billion;
        }

        tp->tv_sec += sec_delta;
        tp->tv_nsec += nsec_delta;
    }

    return ret;
}