package storage

import (
	"database/sql"
	"fmt"
	"sync"
	// 为了加载mysql驱动的init方法
	_ "github.com/go-sql-driver/mysql"
)

type Storage struct {
	db            map[string]*DB
	maxopencounts int
	output        func(string, ...interface{})
}

func New(dbs map[string]string, maxOpenConns int) *Storage {
	if dbs == nil {
		dbs = make(map[string]string)
	}
	storage := &Storage{
		db:            make(map[string]*DB),
		maxopencounts: maxOpenConns,
	}
	for dbname, dsn := range dbs {
		if erro := storage.AddDB(dbname, dsn); erro != nil {
			panic(erro.Error())
		}
	}
	return storage
}

func (s *Storage) AddDB(name string, dsn string) error {
	if _, has := s.db[name]; has {
		return fmt.Errorf("db:%s is already exist", name)
	}
	//dns := fmt.Sprintf("%s/%s?charset=utf8", s.address, name)
	db, erro := sql.Open("mysql", dsn)
	if erro != nil {
		return erro
	}
	db.SetMaxOpenConns(s.maxopencounts)
	db.SetMaxIdleConns(s.maxopencounts)
	s.db[name] = &DB{
		db:     db,
		dbname: name,
		root:   s,
		pool: sync.Pool{New: func() interface{} {
			return &Session{}
		}},
	}
	return nil
}

func (s *Storage) SetLogOutput(output func(string, ...interface{})) {
	s.output = output
}

func (s *Storage) GetDB(name string) *DB {
	if db, has := s.db[name]; has {
		return db
	}
	return nil
}
