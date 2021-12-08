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
#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>

#define THREAD_NUM 1000

pthread_t threads[THREAD_NUM];

void *print_thread( void* thread_id ) {
  printf("Thread %d\n", *((int*)thread_id));

  return NULL;
}

int main() {
  int  index = 0;
  while(1) {
    if (index >= THREAD_NUM) {
      pthread_t last_thread = threads[index % THREAD_NUM];
      pthread_join(last_thread, NULL);
    }
    int now = index++;
    pthread_create( &threads[now % THREAD_NUM], NULL, print_thread, (void*) &now);
  }

  exit(0);
}