package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
)

// DB 封装了对mysql访问的基本操作
type DB struct {
	db     *sql.DB
	pool   sync.Pool
	root   *Storage
	dbname string
}

// Query 查询语句
func (d *DB) Query(query ...string) ([]Row, error) {
	if len(query) == 0 {
		return nil, fmt.Errorf("query func args is empty")
	}
	sqlstr := query[0]
	if len(query) > 1 {
		sqlstr = "/*" + query[1] + "*/ " + sqlstr
	}
	if d.root.output != nil {
		logstr := fmt.Sprintf("[storage] sql(%s)->%s", d.dbname, sqlstr)
		d.root.output(logstr)
	}
	var result = make([]Row, 0)
	rows, erro := d.db.Query(sqlstr)
	if erro != nil {
		if erro == sql.ErrNoRows {
			return result, nil
		}
		return nil, erro
	}
	defer rows.Close()
	// 获取列字段信息
	columns, _ := rows.Columns()
	colnum := len(columns)
	// 构造scanArgs、values两个数组，scanArgs的每个值指向values相应值的地址
	scanArgs, values := make([]interface{}, colnum), make([][]byte, colnum)
	for k := range scanArgs {
		scanArgs[k] = &values[k]
	}
	for rows.Next() {
		if erro := rows.Scan(scanArgs...); erro != nil {
			return nil, erro
		}
		//将行数据保存到record字典
		var record = make(Row)
		for k, v := range values {
			record[columns[k]] = v
		}
		result = append(result, record)
	}
	return result, nil
}

// Exec 执行语句
func (d *DB) Exec(query ...string) (sql.Result, error) {
	if len(query) == 0 {
		return nil, fmt.Errorf("Exec func args is empty")
	}
	sqlstr := query[0]
	if len(query) > 1 {
		sqlstr = "/*" + query[1] + "*/ " + sqlstr
	}
	if d.root.output != nil {
		logstr := fmt.Sprintf("[storage] sql(%s)->%s", d.dbname, sqlstr)
		d.root.output(logstr)
	}
	return d.db.Exec(strings.TrimSpace(sqlstr))
}

// Session 单次查询执行的会话
type Session struct {
	root     *DB
	comments string
	table    string
	fields   string
	where    string
}

func (d *DB) Table(name string) *Session {
	s := d.pool.Get().(*Session)
	s.reset()
	s.table = name
	s.root = d
	return s
}

func (s *Session) reset() {
	s.table = ""
	s.fields = ""
	s.where = ""
	s.comments = ""
}

// Close 用于吧session放回池子中
func (s *Session) Close() {
	if s.root != nil {
		s.root.pool.Put(s)
	}
}

// SetComments 设置注释
func (s *Session) SetComments(notes string) *Session {
	s.comments = notes
	return s
}

// Fields Select 查询参数
func (s *Session) Fields(args ...string) *Session {
	fiedls := strings.Join(args, ",")
	s.fields = fiedls
	return s
}
func (s *Session) Where(str string, args ...interface{}) *Session {
	query := parseSQLArgs(strings.TrimSpace(str), args...)
	s.where = query
	return s
}

func (s *Session) Select(args ...string) ([]Row, error) {
	defer s.Close()
	if len(s.fields) == 0 {
		s.fields = "*"
	}
	sql := fmt.Sprintf("select %s from %s", s.fields, s.table)
	if len(s.where) != 0 {
		sql += " where " + s.where
	}
	if len(args) > 0 {
		sql += " " + strings.Join(args, " ")
	}
	return s.root.Query(sql, s.comments)
}

// Insert 返回插入的主键 当主键是自增长的整数时有效
func (s *Session) Insert(args map[string]interface{}) (int64, error) {
	defer s.Close()
	if len(args) == 0 {
		return 0, fmt.Errorf("insert args is empty")
	}
	colnumsvalue := parseMapToInsert(args)
	sql := fmt.Sprintf(
		"insert into %s %s",
		s.table, colnumsvalue,
	)
	if ret, erro := s.root.Exec(sql, s.comments); erro != nil {
		return 0, erro
	} else {
		lastid, _ := ret.LastInsertId()
		return lastid, nil
	}
}

func (s *Session) Update(args map[string]interface{}) (int64, error) {
	defer s.Close()
	if len(s.where) == 0 {
		return 0, fmt.Errorf("update sql is not found where")
	}
	updatevalue := parseMapToUpdate(args)
	sql := fmt.Sprintf(
		"update %s set %s where %s",
		s.table, updatevalue, s.where,
	)
	if ret, erro := s.root.Exec(sql, s.comments); erro != nil {
		return 0, erro
	} else {
		count, _ := ret.RowsAffected()
		return count, nil
	}
}

func (s *Session) InsertDup(val, update map[string]interface{}) (int64, int64, error) {
	defer s.Close()
	if len(val) == 0 || len(update) == 0 {
		return 0, 0, fmt.Errorf("insertdup args is empty")
	}
	colnumsvalue := parseMapToInsert(val)
	updatevalue := parseMapToUpdate(update)
	sql := fmt.Sprintf(
		"insert into %s %s on duplicate key update %s",
		s.table, colnumsvalue, updatevalue,
	)
	if ret, erro := s.root.Exec(sql, s.comments); erro != nil {
		return 0, 0, erro
	} else {
		inserid, _ := ret.LastInsertId()
		count, _ := ret.RowsAffected()
		return inserid, count, nil
	}

}

func (s *Session) Delete() error {
	defer s.Close()
	if len(s.where) == 0 {
		return fmt.Errorf("delete sql is not found where")
	}
	sql := fmt.Sprintf("delete from %s where %s", s.table, s.where)
	_, erro := s.root.Exec(sql, s.comments)
	return erro
}
