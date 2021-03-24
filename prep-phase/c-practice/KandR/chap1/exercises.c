#include <stdio.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>



#define MAXSIZE 1000;

void detab(int n);
void entab(int n);
int strlen2(char *str);

int main() {
    char *t = "does this" " work";

    int count = 0;

    for ( int i = 0;t[i] != 0; i++) {
        count++;
    }

    printf("count = %d\n", count);

    char *teststr = "mom";

    printf("strlen2 mom = %d\n", strlen2(teststr));

    //printf("%s",t);
    //c = getch();
    //printf("%d\n", EOF);
    //c = (getchar() != c);
    //int t = (int)'\0';
    //printf("%d\n", t);
    //printf("%d\n", EOF);
    // while((c =  getchar()) != EOF) {
    //     if (c == '\t') {
    //         putchar('\\');
    //         putchar('t');
    //     } else if (c == '\b') {
    //         putchar(c);
    //         //putchar('\\');
    //         //putchar('b');
    //     } else if (c == '\\') {
    //         putchar('\\');
    //         putchar('\\');
    //     } else {
    //         putchar(c);
    //     }
        
    // }

    struct stat buf;

    if (access("./exercises.c", F_OK) >= 0) {
        /* file exists */
        printf("file exists\n");
    } 


    if (access("./exercises.c", R_OK) == 0) {
        printf("readable\n");
    }

    /* returns NULL on error */
    FILE *f = fopen("./exercises.c", "r");

    char line[BUFSIZ];

    if (fgets(line, BUFSIZ, f) == NULL) {
        if (ferror(stdin)) {
            perror("getline err");
        }
        else if (feof(stdin)) {
            fprintf(stderr, "end of file\n");
        }

    } else {
        if ('\n' == line[strlen(line) - 1]) {

            /* use line here */
            //printf(line);
            for (int i = 0; i < strlen(line) ;i++) {
                putchar(line[i]);
            }

        } else {
            fprintf(stderr, "long line truncated\n");
        }
    }

    //entab(8);
        
    return 0;
}


void detab(int n) {
    int c;
    int col = 0;
    int totabstop;
    while((c =  getchar()) != EOF) {
        if (c == '\t') {
            totabstop = (col % n) == 0 ? n : n - (col % n);
            col += totabstop;
            while(totabstop > 0) {
                putchar(' ');
                totabstop--;
            }
        } else if (c == '\n') {
            putchar(c);
            col = 0;
        } else {
            putchar(c);
            col++;
        }
    }
}

void entab(int n) {
    int c;
    int col = 0;
    int totabstop;
    int spacecount = 1;
    while((c =  getchar()) != EOF) {
        if (c == ' ') {

            totabstop = (col % n) == 0 ? n : n - (col % n);


            while ((c =  getchar()) == ' ')
                spacecount++;
            col += spacecount + 1;

            
            if (spacecount >= totabstop ) {
                putchar('\t');
                spacecount -= totabstop;
                int tabcount = spacecount / n;

                while (tabcount > 0) {
                    putchar('\t');
                    tabcount--;
                }

                spacecount %= n;

            } 

            while (spacecount > 0) {
                putchar(' ');
                spacecount--;
            }

            putchar(c);
            spacecount = 1;
        } else if (c == '\n') {
            putchar(c);
            col = 0;
        } else {
            putchar(c);
            col++;
        }
    }
}

int strlen2(char *str) {
    //int c;
    int count = 0;
    for (; *str != '\0'; str++)
        count++;
    return count;
}