package levelwrapper

/*
#cgo CFLAGS: -I/opt/homebrew/opt/leveldb/include
#cgo LDFLAGS: -L/opt/homebrew/opt/leveldb/lib -lleveldb
#include <leveldb/c.h>
#include <stdlib.h>
void ldbtest();

void lvldbfree(leveldb_t* db) { free(db); }
void lvldb_optfree(leveldb_options_t *options) { free(options); }
void lvldb_namefree(char* name) { free(name); }
void lvldb_errptrfree(char* errptr) { free(errptr); }
*/
import "C"

import (
  "errors"
  //"unsafe"
)

type Lvldb struct {
  db *C.leveldb_t
  options *C.leveldb_options_t
  name *C.char
  errptr *C.char
}


func Testffi() {
  C.ldbtest()
}


func LevelDBOpen(options *C.leveldb_options_t, name *C.char) (Lvldb, error) {
  var errptr *C.char
  //defer C.free(unsafe.Pointer(errptr))
  db := Lvldb{db:C.leveldb_open(options, name, &errptr), options:options, name: name, errptr: errptr}
  errstr := C.GoString(errptr)
  var err error = nil
  if errstr != "" {
    err = errors.New(errstr)
    return db, err
  }
  return db, nil
  
}

func (db *Lvldb) Free() {
  C.lvldbfree(db.db)
  C.lvldb_optfree(db.options)
  C.lvldb_namefree(db.name)
  C.lvldb_errptrfree(db.errptr)
}


func (db *Lvldb) Close() {
  C.leveldb_close(db.db)
}

func (db *Lvldb) Put(
