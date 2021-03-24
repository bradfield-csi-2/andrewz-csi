#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <termios.h>

int main(int argc, char *argv[], char *envp[]) {
  struct termios term, term_orig;

  if(tcgetattr(0, &term_orig)) {
    printf("tcgetattr failed\n");
    exit(-1);
  }

  term = term_orig;

  term.c_lflag &= ~ICANON;
  term.c_lflag |= ECHO;
  term.c_cc[VMIN] = 0;
  term.c_cc[VTIME] = 0;

  if (tcsetattr(0, TCSANOW, &term)) {
    printf("tcsetattr failed\n");
    exit(-1);
  }

  char ch;
  while (1) {
    if (read(0, &ch, 1) > 0) 
      printf(" %d\n", ch);
  }
  return 0;
}