package main

import (
	"fmt"
	//"os"
)

type tblMgr struct {
	db   string
	root string
}

func initTblMgr(db string, root string) tblMgr {
	return tblMgr{db, root}
}

func (t *tblMgr) GetTableFilepath(tablename string) string {
	filepath := fmt.Sprintf("%s/%s/tables/%s/data", t.root, t.db, tablename)
	return filepath
}

func (t *tblMgr) GetTableSchemaFilepath(tablename string) string {
	filepath := fmt.Sprintf("%s/%s/tables/%s/schema", t.root, t.db, tablename)
	return filepath
}

//get table -- ./testdb/tables/[tableName]/table
//get schema -- ./testdb/tables/[tableName]/schema (meta?)

//get table block
