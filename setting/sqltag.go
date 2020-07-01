package setting

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// StructTag named sql or xorm
type SqlTag struct {
	data    map[string]string
	changed bool
	lock    sync.RWMutex
	reflect.StructTag
}

func NewSqlTag() *SqlTag {
	return &SqlTag{data: make(map[string]string)}
}

func (st *SqlTag) ParseTag(tag reflect.StructTag) {
	st.StructTag, st.changed = tag, true
	str := fmt.Sprintf("%s;%s", tag.Get("sql"), tag.Get(XORM_TAG_NAME))
	pieces := strings.Split(str, ";")
	for _, value := range pieces {
		if strings.TrimSpace(value) == "" {
			continue
		}
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToLower(v[0]))
		if len(v) >= 2 {
			st.data[k] = strings.Join(v[1:], ":")
		} else {
			st.data[k] = k
		}
	}
}

// 转为字符串格式，头通常使用sql或xorm
func (st *SqlTag) String(head string) string {
	if !st.changed {
		return string(st.StructTag)
	}
	var pairs []string
	for key, val := range st.data {
		if val == "" || val == key {
			pairs = append(pairs, fmt.Sprintf("%s", key))
		} else {
			pairs = append(pairs, fmt.Sprintf("%s:%s", key, val))
		}
	}
	var result string
	if len(pairs) > 0 {
		result = fmt.Sprintf(`%s:"%s"`, head, strings.Join(pairs, ";"))
	}
	st.StructTag = reflect.StructTag(result)
	return result
}

// Returns a tag from the tag data
func (st *SqlTag) Get(key string) (string, bool) {
	st.lock.RLock()
	defer st.lock.RUnlock()
	val, ok := st.data[key]
	return val, ok
}

// Sets a tag in the tag data map
func (st *SqlTag) Set(key, val string) {
	st.lock.Lock()
	defer st.lock.Unlock()
	st.data[key] = val
	st.changed = true
}

// Deletes a tag
func (st *SqlTag) Delete(key string) {
	st.lock.Lock()
	defer st.lock.Unlock()
	delete(st.data, key)
	st.changed = true
}
