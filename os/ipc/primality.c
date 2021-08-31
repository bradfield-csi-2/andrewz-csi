#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>

int brute_force(long n);
int brutish(long n);
int miller_rabin(long n);

void exit_with_usage() {
  //fprintf(stderr, "Usage: ./primality [brute_force|brutish|miller_rabin] [-r rangeStart rangeEnd ]\n");
  fprintf(stderr, "Usage: ./primality [brute_force|brutish|miller_rabin] \n");

  exit(1);
}

int main(int argc, char*argv[]) {
  long num;
  int (*func)(long), tty;

  if (argc != 2)
    exit_with_usage();
  //alllow more than two
  //have a flag for -r range start end
  //do all in range for loop

  if (strcmp(argv[1], "brute_force") == 0)
    func = &brute_force;
  else if (strcmp(argv[1], "brutish") == 0)
    func = &brutish;
  else if (strcmp(argv[1], "miller_rabin") == 0)
    func = &miller_rabin;
  else
    exit_with_usage();


  //fprintf(stderr, "Called primality with %d args. And method %s \n", argc, argv[1]);

  tty = isatty(fileno(stdin));
  /*
  int rangeFlag = 0;
  if (argc == 5 && strcmp(argv[2], "-r") == 0)
    rangeFlag = 1;
  else if (argc == 5 && strcmp(argv[2], "-r") != 0)
    exit_with_usage();

  fprintf(stderr, "Range flag with method %s is %s \n", argv[1], rangeFlag ? "ON" : "OFF");





  if (rangeFlag) {
    char *eptr;
    long rstart, rend;

    rstart = strtol(argv[3], &eptr, 0);

    // If the result is 0, test for an error 
    if (rstart == 0)
    {
        // If a conversion error occurred, display a message and exit 
        if (errno == EINVAL)
        {
            //fprintf(stderr, "Conversion error occurred for range start: %d\n", errno);
            perror("Conversion error occurred for range start");
            exit(0);
        }
    }


    rend = strtol(argv[4], &eptr, 0);

    // If the result is 0, test for an error 
    if (rend == 0)
    {
        // If a conversion error occurred, display a message and exit 
        if (errno == EINVAL)
        {
            //fprintf(stderr,"Conversion error occurred for range end: %d\n", errno);
            perror("Conversion error occurred for range end");
            exit(0);
        }
    }

    

    fprintf(stderr, "Range \"%s\", start: %ld || end: %ld :\n> ", argv[1], rstart, rend);
    //perror(sprintf("Range \"%s\", start: %ld || end: %ld :\n> ", argv[1], rstart, rend));

    if (tty) {

      

      for (num = rstart ; num <= rend; num++ ) {
        //read(STDIN_FILENO, &num, sizeof(num));
        //iint result = (*func)(num);
        printf("%ld: %d\n",num, (*func)(num));
        fprintf(stderr, "Doing work in one\n");
        //write(STDOUT_FILENO, &result, sizeof(result));
      }
    } else {
      //
      for (num = rstart ; num <= rend; num++ ) {
        //read(STDIN_FILENO, &num, sizeof(num));
        int result = (*func)(num);
        //printf("%ld: %d\n",num, (*func)(num));
        //if ( write(STDOUT_FILENO, &result, sizeof(result)) == -1 )
          //perror("Error on write");


        
        if (result != 0) {
          if ( write(STDOUT_FILENO, &num, sizeof(num)) == -1 )
            perror("Error on write");
        }
        
      }
      num = -1;
      if ( write(STDOUT_FILENO, &num, sizeof(num)) == -1 )
        perror("Error on write");
 
      //write(STDOUT_FILENO, &num, sizeof(num));
 
    }
    exit(0);
  }

    */
  if (tty) {

    fprintf(stderr, "Running \"%s\", enter a number:\n> ", argv[1]);

    while (scanf("%ld", &num) == 1) {
      printf("%d\n", (*func)(num));
      fflush(stdout);
      fprintf(stderr, "> ");
    }
  } else {
    for (;;) {
      read(STDIN_FILENO, &num, sizeof(num));
      int result = (*func)(num);
      write(STDOUT_FILENO, &result, sizeof(result));
    }
  }
}

/*
 * Primality test implementations
 */

// Just test every factor
int brute_force(long n) {
  for (long i = 2; i < n; i++)
    if (n % i == 0)
      return 0;
  return 1;
}

// Test factors, up to sqrt(n)
int brutish(long n) {
  long max = floor(sqrt(n));
  for (long i = 2; i <= max; i++)
    if (n % i == 0)
      return 0;
  return 1;
}

int randint(int a, int b) { return rand() % (++b - a) + a; }

int modpow(int a, int d, int m) {
  int c = a;
  for (int i = 1; i < d; i++)
    c = (c * a) % m;
  return c % m;
}

int witness(int a, int s, int d, int n) {
  int x = modpow(a, d, n);
  if (x == 1)
    return 1;
  for (int i = 0; i < s - 1; i++) {
    if (x == n - 1)
      return 1;
    x = modpow(x, 2, n);
  }
  return (x == n - 1);
}

// TODO we should probably make this a parameter!
int MILLER_RABIN_ITERATIONS = 10;

// An implementation of the probabilistic Miller-Rabin test
int miller_rabin(long n) {
  int a, s = 0, d = n - 1;

  if (n == 2)
    return 1;

  if (!(n & 1) || n <= 1)
    return 0;

  while (!(d & 1)) {
    d >>= 1;
    s += 1;
  }
  for (int i = 0; i < MILLER_RABIN_ITERATIONS; i++) {
    a = randint(2, n - 1);
    if (!witness(a, s, d, n))
      return 0;
  }
  return 1;
}
