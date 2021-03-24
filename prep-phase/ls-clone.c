//ls clone
//takes a file, default is current directory

//Minimally, it should list the contents of a 
//directory including some information about each file, 
//such as file size. As a stretch goal, use man ls to identify 
//any interesting flags you may wish to support, andimplement them.


#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <dirent.h>



struct stat buf;

struct stat dirEntBuf;

const int MAXPATHLEN = 100;

const char *FILE_TYPE = "File";

const char *DIR_TYPE = "Directory";

struct fileInfoEntry {
    char fileName[32];
    off_t filesize;
};

int countDirEntries(char * dirPath);
struct fileInfoEntry *getFiles(char *dirName, int dirFileCount);

void printDirFiles(struct fileInfoEntry *dirFiles, int dirFileCount);

int main(int argc, char **argv) {
    
    char *dirname;

    if (argc == 2) {
        dirname = calloc(80, sizeof(char));
        sprintf(dirname,"%s/",argv[1]) ;
    } else {
        dirname = "./";
    }

    int dirCount = countDirEntries(dirname);
    struct fileInfoEntry *files = getFiles(dirname, dirCount);

    printDirFiles(files, dirCount);


    


    return 0;
}

struct fileInfoEntry *getFiles(char *dirName, int dirFileCount){
    
    int dirNameLen = strlen(dirName);
    int largestStrLen = dirNameLen + 10;
    int len;
    char *filePath  = (char *) calloc(largestStrLen, sizeof(*filePath));
	struct fileInfoEntry *fileEntry;
    struct fileInfoEntry *dirFiles;
	int fileIndex = 0;	
    
	dirFiles = calloc(dirFileCount, sizeof(*dirFiles));

    if (stat(dirName, &buf) != 0) {
        perror("stat failed");
    } else if (S_ISDIR(buf.st_mode)) {
        
        DIR *dir = opendir(dirName);
        struct dirent *de;

        while ((de = readdir(dir))) {

            len = dirNameLen + strlen(de->d_name) + 1;
            
            if (len > largestStrLen) {
                largestStrLen = len;
                free(filePath);
                filePath = (char *) calloc(len, sizeof(*filePath));

            }
            
            strcpy(filePath, dirName);
            strcat(filePath, de->d_name);

	        fileEntry = malloc(sizeof(fileEntry));

            if (stat(filePath, &dirEntBuf) != 0) {
                perror("dir ent stat failed");
                strcpy(fileEntry->fileName,"Error");
                fileEntry->filesize = 0;
            } else if (S_ISREG(dirEntBuf.st_mode)) {
                strcpy(fileEntry->fileName,de->d_name);
                fileEntry->filesize = dirEntBuf.st_size;
            } else if (S_ISDIR(dirEntBuf.st_mode)) {
                strcpy(fileEntry->fileName,de->d_name);
                fileEntry->filesize = dirEntBuf.st_size;
            } else {
                perror("unhandled file type");
                strcpy(fileEntry->fileName,"Unhandled");
                fileEntry->filesize = 0;
            }

            dirFiles[fileIndex] = *fileEntry;	
            fileIndex++;
        }

        free(filePath);
        closedir(dir);

    } else {
        perror("Invalid, can't get files on non directory");
    }

   return dirFiles; 
}


int countDirEntries(char * dirPath) {

    int dirCount = 0;

    if (stat(dirPath, &buf) != 0) {
        perror("stat failed");
    } else if (S_ISDIR(buf.st_mode)) {

        DIR *dir = opendir(dirPath);
        struct dirent *de;

        while ((de = readdir(dir))) {
		    dirCount++;
	    }
    } else {
	    perror("unhandled file type, not a directory");
    }

    return dirCount;
}	


void printDirFiles(struct fileInfoEntry *dirFiles, int dirFileCount) {
    
    printf("HEADER\n");
    
    for (;dirFileCount > 0; dirFiles++, dirFileCount--) {
        
        printf("name: %s || size in bytes: %lld\n", dirFiles->fileName, dirFiles->filesize);
        //dirFiles++;
        //dirFileCount--;
    }
}

