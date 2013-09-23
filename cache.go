package xorm

import (
	"container/list"
	//"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type CacheStore interface {
	Put(key, value interface{}) error
	Get(key interface{}) (interface{}, error)
	Del(key interface{}) error
}

type MemoryStore struct {
	store map[interface{}]interface{}
	mutex sync.Mutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{store: make(map[interface{}]interface{})}
}

func (s *MemoryStore) Put(key, value interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[key] = value
	//fmt.Println("after put store:", s.store)
	return nil
}

func (s *MemoryStore) Get(key interface{}) (interface{}, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	//fmt.Println("before get store:", s.store)
	if v, ok := s.store[key]; ok {
		return v, nil
	}

	return nil, ErrNotExist
}

func (s *MemoryStore) Del(key interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	//fmt.Println("before del store:", s.store)
	delete(s.store, key)
	//fmt.Println("after del store:", s.store)
	return nil
}

type Cacher interface {
	GetIds(tableName, sql string) interface{}
	GetBean(tableName string, id int64) interface{}
	PutIds(tableName, sql string, ids interface{})
	PutBean(tableName string, id int64, obj interface{})
	DelIds(tableName, sql string)
	DelBean(tableName string, id int64)
	ClearIds(tableName string)
}

// LRUCacher implements Cacher according to LRU algorithm
type LRUCacher struct {
	idList   *list.List
	sqlList  *list.List
	idIndex  map[interface{}]*list.Element
	sqlIndex map[string]map[interface{}]*list.Element
	store    CacheStore
	Max      int
	mutex    sync.Mutex
}

func NewLRUCacher(store CacheStore, max int) *LRUCacher {
	cacher := &LRUCacher{store: store, idList: list.New(),
		sqlList: list.New(), idIndex: make(map[interface{}]*list.Element),
		Max: max}
	cacher.sqlIndex = make(map[string]map[interface{}]*list.Element)
	return cacher
}

func (m *LRUCacher) GetIds(tableName, sql string) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if v, err := m.store.Get(sql); err == nil {
		if _, ok := m.sqlIndex[tableName]; !ok {
			m.sqlIndex[tableName] = make(map[interface{}]*list.Element)
		}
		if el, ok := m.sqlIndex[tableName][sql]; !ok {
			el = m.sqlList.PushBack(sql)
			m.sqlIndex[tableName][sql] = el
		} else {
			m.sqlList.MoveToBack(el)
		}
		return v
	}
	if tel, ok := m.sqlIndex[tableName]; ok {
		if el, ok := tel[sql]; ok {
			delete(m.sqlIndex[tableName], sql)
			m.sqlList.Remove(el)
		}
	}
	return nil
}

func (m *LRUCacher) GetBean(tableName string, id int64) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	tid := genId(tableName, id)
	if v, err := m.store.Get(tid); err == nil {
		if el, ok := m.idIndex[tid]; ok {
			m.idList.MoveToBack(el)
		} else {
			el = m.idList.PushBack(tid)
			m.idIndex[tid] = el
		}
		return v
	}
	if el, ok := m.idIndex[tid]; ok {
		delete(m.idIndex, tid)
		m.idList.Remove(el)
		if ms, ok := m.sqlIndex[tableName]; ok {
			for _, v := range ms {
				m.sqlList.Remove(v)
			}
			m.sqlIndex[tableName] = make(map[interface{}]*list.Element)
		}
	}
	return nil
}

func (m *LRUCacher) PutIds(tableName, sql string, ids interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.sqlIndex[tableName]; !ok {
		m.sqlIndex[tableName] = make(map[interface{}]*list.Element)
	}
	if el, ok := m.sqlIndex[tableName][sql]; !ok {
		el = m.sqlList.PushBack(sql)
		m.sqlIndex[tableName][sql] = el
	}
	m.store.Put(sql, ids)
}

func (m *LRUCacher) PutBean(tableName string, id int64, obj interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var el *list.Element
	var ok bool
	tid := genId(tableName, id)
	if el, ok = m.idIndex[tid]; !ok {
		el = m.idList.PushBack(tid)
		m.idIndex[tid] = el
	}

	m.store.Put(tid, obj)
	if m.idList.Len() > m.Max {
		e := m.idList.Front()
		m.store.Del(e.Value)
		delete(m.idIndex, e.Value)
		m.idList.Remove(e)
	}
}

func (m *LRUCacher) DelIds(tableName, sql string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.sqlIndex[tableName]; ok {
		if el, ok := m.sqlIndex[tableName][sql]; ok {
			m.store.Del(sql)
			delete(m.sqlIndex, sql)
			m.sqlList.Remove(el)
		}
	}
}

func (m *LRUCacher) DelBean(tableName string, id int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	tid := genId(tableName, id)
	if el, ok := m.idIndex[tid]; ok {
		m.store.Del(tid)
		delete(m.idIndex, tid)
		m.idList.Remove(el)
		if tis, ok := m.sqlIndex[tableName]; ok {
			for _, v := range tis {
				m.sqlList.Remove(v)
			}
			m.sqlIndex[tableName] = make(map[interface{}]*list.Element)
		}
	}
}

func (m *LRUCacher) ClearIds(tableName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if tis, ok := m.sqlIndex[tableName]; ok {
		for _, v := range tis {
			m.sqlList.Remove(v)
		}
		m.sqlIndex[tableName] = make(map[interface{}]*list.Element)
	}
}

func encodeIds(ids []int64) (s string) {
	s = "["
	for _, id := range ids {
		s += fmt.Sprintf("%v,", id)
	}
	s = s[:len(s)-1] + "]"
	return
}

func decodeIds(s string) []int64 {
	res := make([]int64, 0)
	if len(s) >= 2 {
		ss := strings.Split(s[1:len(s)-1], ",")
		for _, s := range ss {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return res
			}
			res = append(res, i)
		}
	}
	return res
}

func getCacheSql(m Cacher, tableName, sql string, args interface{}) ([]int64, error) {
	bytes := m.GetIds(tableName, genSqlKey(sql, args))
	if bytes == nil {
		return nil, errors.New("Not Exist")
	}
	objs := decodeIds(bytes.(string))
	return objs, nil
}

func putCacheSql(m Cacher, ids []int64, tableName, sql string, args interface{}) error {
	bytes := encodeIds(ids)
	m.PutIds(tableName, genSqlKey(sql, args), bytes)
	return nil
}

func genSqlKey(sql string, args interface{}) string {
	return fmt.Sprintf("%v-%v", sql, args)
}

func genId(prefix string, id int64) string {
	return fmt.Sprintf("%v-%v", prefix, id)
}
