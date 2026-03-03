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
#include <sys/time.h>
#include <inttypes.h>
#include <syscall.h>
#include <math.h>

extern int64_t TV_SEC_DELTA;
extern int64_t TV_NSEC_DELTA;

#if defined(__amd64__)
inline int real_gettimeofday(struct timeval *tv, struct timezone *tz)
{
    int ret;
    asm volatile(
        "syscall"
        : "=a"(ret)
        : "0"(__NR_gettimeofday), "D"(tv), "S"(tz)
        : "memory");

    return ret;
}

#elif defined(__aarch64__)
inline int real_gettimeofday(struct timeval *tv, struct timezone *tz)
{
    register int w0 __asm__("w0");

    register struct timeval *x0 __asm__("x0") = tv;
    register struct timezone *x1 __asm__("x1") = tz;
    register uint64_t w8 __asm__("w8") = __NR_gettimeofday; /* syscall number */
    __asm__ __volatile__(
        "svc 0;"
        : "+r"(w0)
        : "r"(x0), "r" (x1), "r"(w8)
        : "memory");

    return w0;
}
#endif

int fake_gettimeofday(struct timeval *tv, struct timezone *tz)
{
    int ret = real_gettimeofday(tv, tz);

    int64_t sec_delta = TV_SEC_DELTA;
    int64_t nsec_delta = TV_NSEC_DELTA;
    int64_t billion = 1000000000;

    while (nsec_delta + tv->tv_usec*1000 > billion)
    {
        sec_delta += 1;
        nsec_delta -= billion;
    }

    while (nsec_delta + tv->tv_usec*1000 < 0)
    {
        sec_delta -= 1;
        nsec_delta += billion;
    }

    tv->tv_sec += sec_delta;
    tv->tv_usec += round(nsec_delta/1000);

    return ret;
}
