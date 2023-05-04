package db

import (
	_ "fmt"

	"database/sql"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/utils"
)

type DataBase struct {
	db     orm.Ormer
	tx     orm.TxOrmer
	dbName string
}

func NewDataBase(dbName string) *DataBase {
	return &DataBase{dbName: dbName}
}

func (this *DataBase) GetOrmer() orm.Ormer {
	return this.db
}

func (this *DataBase) GetTxOrmer() orm.TxOrmer {
	return this.tx
}

func (this *DataBase) Read(entity interface{}, cols ...string) error {
	if this.tx != nil {
		return this.tx.Read(entity, cols...)
	} else {
		return this.db.Read(entity, cols...)
	}
}
func (this *DataBase) ReadOrCreate(entity interface{}, col string, cols ...string) (bool, int64, error) {
	if this.tx != nil {
		return this.tx.ReadOrCreate(entity, col, cols...)
	} else {
		return this.db.ReadOrCreate(entity, col, cols...)
	}
}
func (this *DataBase) Insert(entity interface{}) (int64, error) {
	if this.tx != nil {
		return this.tx.Insert(entity)
	} else {
		return this.db.Insert(entity)
	}
}
func (this *DataBase) InsertMulti(bulk int, entities interface{}) (int64, error) {
	if this.tx != nil {
		return this.tx.InsertMulti(bulk, entities)
	} else {
		return this.db.InsertMulti(bulk, entities)
	}
}
func (this *DataBase) Update(entity interface{}, colConflictAndArgs ...string) (int64, error) {
	if this.tx != nil {
		return this.tx.Update(entity, colConflictAndArgs...)
	} else {
		return this.db.Update(entity, colConflictAndArgs...)
	}
}
func (this *DataBase) Delete(entity interface{}) (int64, error) {
	if this.tx != nil {
		return this.tx.Delete(entity)
	} else {
		return this.db.Delete(entity)
	}
}
func (this *DataBase) LoadRelated(entity interface{}, name string, args ...utils.KV) (int64, error) {
	if this.tx != nil {
		return this.tx.LoadRelated(entity, name, args...)
	} else {
		return this.db.LoadRelated(entity, name, args...)
	}
}
func (this *DataBase) QueryM2M(entity interface{}, name string) orm.QueryM2Mer {
	if this.tx != nil {
		return this.tx.QueryM2M(entity, name)
	} else {
		return this.db.QueryM2M(entity, name)
	}
}
func (this *DataBase) QueryTable(ptrStructOrTableName interface{}) orm.QuerySeter {
	if this.tx != nil {
		return this.tx.QueryTable(ptrStructOrTableName)
	} else {
		return this.db.QueryTable(ptrStructOrTableName)
	}
}
func (this *DataBase) Open() {
	this.db = orm.NewOrmUsingDB(this.dbName)
}

func (this *DataBase) Begin() (err error) {
	this.Open()
	this.tx, err = this.db.Begin()
	return err
}

func (this *DataBase) BeginWithOpts(opts *sql.TxOptions) (err error) {
	this.Open()
	this.tx, err = this.db.BeginWithOpts(opts)
	return err
}

func (this *DataBase) Commit() (err error) {
	if this.tx != nil {
		return this.tx.Commit()
	}
	return err
}

func (this *DataBase) Rollback() (err error) {
	if this.tx != nil {
		return this.tx.Rollback()
	}
	return err
}

func (this *DataBase) Raw(query string, args ...interface{}) orm.RawSeter {

	if this.tx != nil {
		return this.tx.Raw(query, args...)
	} else {
		return this.db.Raw(query, args...)
	}

}
