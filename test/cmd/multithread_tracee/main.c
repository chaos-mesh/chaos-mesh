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