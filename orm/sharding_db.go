package orm

import "database/sql"

type ShardingDB struct {
	// key 就是 Dst 里面的 DB
	DBs map[string]*MasterSlavesDB
}

type MasterSlavesDB struct {
	Master *sql.DB
	//Table []string
	Slaves []*sql.DB
}
