#include <leveldb/c.h>
#include <stdlib.h>
#include <stdio.h>
#include <stdbool.h>

void ldbtest(void)
{
  char *errptr = malloc((size_t)400);
  leveldb_options_t *options  = leveldb_options_create();
  leveldb_options_set_create_if_missing(options , true);
  leveldb_t *db = leveldb_open(options, "/tmp/testdb", &errptr);
  //int x = 0;

  //leveldb_writeoptions_t *wopts = leveldb_writeoptions_create();
  if (*errptr) {
    printf("%s\n",errptr);
  }

  leveldb_close(db);

  leveldb_destroy_db(options, "/tmp/testdb", &errptr);

  if (*errptr) {
    printf("%s\n",errptr);
  }

}
