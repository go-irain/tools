package storage

import (
	"database/sql"
	"fmt"
	"sync"
	// 为了加载mysql驱动的init方法
	_ "github.com/go-sql-driver/mysql"
)

type Storage struct {
	address       string
	db            map[string]*DB
	maxopencounts int
	output        func(string, ...interface{})
}

func New(address string, maxOpenConns int) *Storage {
	s := &Storage{
		address:       address,
		maxopencounts: maxOpenConns,
		db:            make(map[string]*DB),
	}
	return s
}

func (s *Storage) AddDB(name string) error {
	if _, has := s.db[name]; has {
		return fmt.Errorf("db:%s is already exist", name)
	}
	dns := fmt.Sprintf("%s/%s?charset=utf8", s.address, name)
	db, erro := sql.Open("mysql", dns)
	if erro != nil {
		return erro
	}
	db.SetMaxOpenConns(s.maxopencounts)
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
