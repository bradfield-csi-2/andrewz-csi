#include <stdio.h>
#include <stdlib.h>


int main() {
  printf("hello\n");
  while(true) {
    malloc((size_t) 1000000);
  }

}
