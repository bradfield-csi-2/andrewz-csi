package db 

/*
Exposes DB interface to kv store
*/


type DB interface {
	Init(filepath string) error
	Get(key string) ([]byte , bool)
	Put(key string, value []byte) //? true false? or ok
	Delete(key string) //true false semantics? 
	//what about transfer protocol? 
	ReadBin(offset int64, buf []byte) (int, error )
}