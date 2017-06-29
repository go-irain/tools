package storage

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Row map[string][]byte

type Express string

func (r Row) Delete(key string) {
	delete(r, key)
}

func (r Row) IsNull(name string) bool {
	if v, has := r[name]; has {
		return v == nil
	}
	return true
}

func (r Row) Str(name string) string {
	if v, has := r[name]; has {
		if v != nil {
			return string(v)
		}
	}
	return ""
}

func (r Row) Int(name string) int64 {
	if v, has := r[name]; has {
		if v != nil {
			num, _ := strconv.ParseInt(string(v), 10, 64)
			return num
		}
	}
	return 0
}

func (r Row) Number(name string) float64 {
	if v, has := r[name]; has {
		if v != nil {
			num, _ := strconv.ParseFloat(string(v), 64)
			return num
		}
	}
	return 0
}

// 修改buff长度
// appendSize buf需要追加的长度
func reserveBuffer(buf []byte, appendSize int) []byte {
	newSize := len(buf) + appendSize
	if cap(buf) < newSize {
		// Grow buffer exponentially
		newBuf := make([]byte, len(buf)*2+appendSize)
		copy(newBuf, buf)
		buf = newBuf
	}
	return buf[:newSize]
}

func escapeStringBackslash(buf []byte, v string) []byte {
	pos := len(buf)
	buf = reserveBuffer(buf, len(v)*2)
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case '\x00':
			buf[pos] = '\\'
			buf[pos+1] = '0'
			pos += 2
		case '\n':
			buf[pos] = '\\'
			buf[pos+1] = 'n'
			pos += 2
		case '\r':
			buf[pos] = '\\'
			buf[pos+1] = 'r'
			pos += 2
		case '\x1a':
			buf[pos] = '\\'
			buf[pos+1] = 'Z'
			pos += 2
		case '\'':
			buf[pos] = '\\'
			buf[pos+1] = '\''
			pos += 2
		case '"':
			buf[pos] = '\\'
			buf[pos+1] = '"'
			pos += 2
		case '\\':
			buf[pos] = '\\'
			buf[pos+1] = '\\'
			pos += 2
		default:
			buf[pos] = c
			pos += 1
		}
	}

	return buf[:pos]
}

func parseArg(result []byte, arg interface{}) []byte {
	switch a := arg.(type) {
	case int:
		result = strconv.AppendInt(result, int64(a), 10)
	case int64:
		result = strconv.AppendInt(result, a, 10)
	case float32:
		result = strconv.AppendFloat(result, float64(a), 'f', -1, 32)
	case float64:
		result = strconv.AppendFloat(result, a, 'f', -1, 64)
	case bool:
		if a {
			result = strconv.AppendInt(result, 1, 10)
		} else {
			result = strconv.AppendInt(result, 0, 10)
		}
	case []byte:
		result = append(result, a...)
	case Express:
		result = append(result, []byte(fmt.Sprintf("%v", a))...)
	default:
		t := reflect.TypeOf(arg)
		if reflect.Slice == t.Kind() {
			result = append(result, ')')
			v := reflect.ValueOf(arg)
			for i := 0; i < v.Len(); i++ {
				result = parseArg(result, v.Index(i).Interface())
				if i != v.Len() {
					result = append(result, ',')
				}
			}
			result = append(result, ')')
		} else {
			result = append(result, '\'')
			result = escapeStringBackslash(result, fmt.Sprintf("%v", a))
			result = append(result, '\'')
		}
	}
	return result
}

func parseSQLArgs(query string, args ...interface{}) string {
	sql := []byte(query)
	result := make([]byte, 0, 0)
	argslen := len(args)
	count := 0
	for _, v := range sql {
		if v == '?' {
			if count == argslen {
				break
			}
			result = parseArg(result, args[count])
			count++
		} else {
			result = append(result, v)
		}
	}
	if count != len(args) {
		panic("?,args count error")
	}
	return string(result)
}

// ParseMapToUpdate 把map转为 key = value , key = value ...
func parseMapToUpdate(args map[string]interface{}) string {
	if len(args) == 0 {
		return ""
	}
	data := make([]byte, 0, 0)
	for k, v := range args {
		key := "`" + k + "`"
		data = append(data, key...)
		data = append(data, '=')
		data = parseArg(data, v)
		data = append(data, ',')
	}
	return string(data[:len(data)-1])
}

// ParseMapToInsert 把map转为 (key1,key2) values (value1,value2)
func parseMapToInsert(args map[string]interface{}) string {
	length := len(args)
	keys := make([]string, 0, 0)
	for k := range args {
		keys = append(keys, "`"+k+"`")
	}
	sort.Strings(keys)
	data := make([]byte, 0, 0)
	data = append(data, '(')
	data = append(data, strings.Join(keys, ",")...)
	data = append(data, ") values ("...)
	for index, field := range keys {
		field = strings.Trim(field, "`")
		data = parseArg(data, args[field])
		if index+1 != length {
			data = append(data, ',')
		}
	}
	data = append(data, ')')
	return string(data)
}
