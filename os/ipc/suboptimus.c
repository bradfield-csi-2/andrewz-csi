#include <signal.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/wait.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <errno.h>
#include <sys/time.h>
#include <sys/select.h>

int START = 2, END = 20;
char *TESTS[] = {"brute_force", "brutish", "miller_rabin"};
int num_tests = sizeof(TESTS) / sizeof(char *);

int main(int argc, char *argv[]) {
  int testfds[num_tests][2];
  int resultfds[num_tests][2];
  int spfds[num_tests][2];
  int activefds[num_tests];
  long currnums[num_tests];
  int result, i;
  long n, rslt;
  char str[10];
  pid_t pid;

  //char sprintf_buffer1[20];
  //sprintf(sprintf_buffer1, "%d", START);
  //char sprintf_buffer2[20];
  //sprintf(sprintf_buffer2, "%d", END);

  if (argc == 3 ) {
     START = atoi(argv[1]);
     /* If the result is 0, test for an error */
     if (START == 0)
     {
         /* If a conversion error occurred, display a message and exit */
         if (errno == EINVAL)
         {
             //fprintf(stderr,"Conversion error occurred for range end: %d\n", errno);
             perror("Conversion error occurred for range start");
             exit(0);
         }
     }
 
     END = atoi(argv[2]);
     /* If the result is 0, test for an error */
     if (END == 0)
     {
         /* If a conversion error occurred, display a message and exit */
         if (errno == EINVAL)
         {
             //fprintf(stderr,"Conversion error occurred for range end: %d\n", errno);
             perror("Conversion error occurred for range end");
             exit(0);
         }
     }
    //sprintf(sprintf_buffer1, "%s", argv[1]);
    //sprintf(sprintf_buffer2, "%s", argv[2]);
  }// else {
    //sprintf(sprintf_buffer1, "%d", START);
    //sprintf(sprintf_buffer2, "%d", END);
  //}

  FILE *log;
  char logFilename[] = "log.txt";

  log = fopen(logFilename, "w");

  if (log == NULL) {
    fprintf(stderr, "Can't open log file %s!\n",
            logFilename);
    exit(1);
  }
  for (i = 0; i < num_tests; i++) {
    //pipe(testfds[i]);
    //pipe(resultfds[i]);

    if (socketpair(AF_UNIX, SOCK_STREAM, 0, spfds[i]) == -1) {
        perror("socketpair");
        exit(1);
    }

    pid = fork();

    if (pid == -1) {
      fprintf(stderr, "Failed to fork\n");
      exit(-1);
    }

    if (pid == 0) {
      dup2(spfds[i][0], STDIN_FILENO);
      dup2(spfds[i][1], STDOUT_FILENO);
      execl("primality", "primality", TESTS[i], (char *)NULL);

    }

    activefds[i] = 1;
    currnums[i] = START;
  }


  struct timespec timeout = {0, 500};
  fd_set readfds;//, writedfs, exceptdfs;

  if (num_tests < 1) {
    exit(0);
  }

  int nfds = spfds[num_tests - 1][0] + 1;
  n = START;
  int ready;

  for (i = 0; i < num_tests; i++) {
    currnums[i] = START;
    send(spfds[i][1], &n, sizeof(n), MSG_DONTWAIT);
  }

 
  //for (n = START; n <= END; n++) {
  int closed_cnt = 0;
  for (;;) {

    FD_ZERO(&readfds);
    for (i = 0; i < num_tests; i++) {
      if ( activefds[i] ) 
        FD_SET(spfds[i][0], &readfds);
    }

    ready = pselect(nfds, &readfds, NULL, NULL,
                &timeout, NULL);

    if ( ready < 1)
      continue;

    for (i = 0; i < num_tests; i++) {

      if (activefds[i] == 0 || FD_ISSET(spfds[i][0], &readfds) == 0 )
        continue;
      if ( read(spfds[i][0], &result, sizeof(result)) == -1 )
       perror("Error on read");

      n = currnums[i];

      fprintf(stdout,"%15s says %ld %s prime\n", TESTS[i], n, result ? "is" : "IS NOT");

      currnums[i] = ++n;

      if( n > END ) {
        activefds[i] = 0;
        closed_cnt++;
        continue;
      }

      write(spfds[i][1], &n, sizeof(n));

    }
    if (closed_cnt == num_tests)
      break;
  }
  fprintf(stdout, "FINISHED\n");
}
