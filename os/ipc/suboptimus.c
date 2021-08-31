#include <signal.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/wait.h>
#include <unistd.h>

int START = 20000, END = 80000;
char *TESTS[] = {"brute_force", "brutish", "miller_rabin"};
int num_tests = sizeof(TESTS) / sizeof(char *);

int main(int argc, char *argv[]) {
  int testfds[num_tests][2];
  int resultfds[num_tests][2];
  int activefds[num_tests];
  int result, i;
  long n, rslt;
  char str[10];
  pid_t pid;

  char sprintf_buffer1[20];
  //sprintf(sprintf_buffer1, "%d", START);
  char sprintf_buffer2[20];
  //sprintf(sprintf_buffer2, "%d", END);

  if (argc == 3 ) {
    sprintf(sprintf_buffer1, "%s", argv[1]);
    sprintf(sprintf_buffer2, "%s", argv[2]);
    //TODO: check for atoi
  } else {
    sprintf(sprintf_buffer1, "%d", START);
    sprintf(sprintf_buffer2, "%d", END);
  }

  for (i = 0; i < num_tests; i++) {
    pipe(testfds[i]);
    pipe(resultfds[i]);

    activefds[i] = 1;

    pid = fork();

    if (pid == -1) {
      fprintf(stderr, "Failed to fork\n");
      exit(-1);
    }

    if (pid == 0) {
      // we are the child, connect the pipes correctly and exec!
      close(testfds[i][1]);
      close(resultfds[i][0]);
      dup2(testfds[i][0], STDIN_FILENO);
      dup2(resultfds[i][1], STDOUT_FILENO);
      execl("primality", "primality", TESTS[i], "-r", sprintf_buffer1, sprintf_buffer2, (char *)NULL);
    }

    // we are the parent
    close(testfds[i][0]);
    close(resultfds[i][1]);
  }

  // for each number, run each test
  //for (n = START; n <= END; n++) {
  int closed_cnt = 0;
  for (;;) {

    for (i = 0; i < num_tests; i++) {

      if (activefds[i] == 0 )
        continue;
      // we are the parent, so send test case to child and read results
      //write(testfds[i][1], &n, sizeof(n));
      //if ( read(resultfds[i][0], &result, sizeof(result)) == -1 )
       //perror("Error on read");
      if ( read(resultfds[i][0], &rslt, sizeof(rslt)) == -1 )
       perror("Error on read");

      if ( rslt == -1 ) {
        activefds[i] = 0;
        closed_cnt++;
        continue;
      }
 
      /*
      printf("%lu is the size of result \n", sizeof(result));
      unsigned char* p = (unsigned char*)&result;
      printf("%x", p[0]); // outputs the first byte of `foo`
      printf("%x", p[1]);
      printf("%x", p[2]);
      printf("%x\n", p[3]);
      */
      printf("%15s says result of %ld is prime \n", TESTS[i], rslt);
      //printf("%15s says %ld %s prime\n", TESTS[i], n, result ? "is" : "IS NOT");
      //printf("%15s says result of %ld is %d \n", TESTS[i], n, result);
    }
    if (closed_cnt == num_tests)
      break;
  }
  fprintf(stdout, "FINISHED\n");
}
