package xorm

import (
	"container/list"
	//"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type CacheStore interface {
	Put(key, value interface{}) error
	Get(key interface{}) (interface{}, error)
	Del(key interface{}) error
}

type MemoryStore struct {
	store map[interface{}]interface{}
	mutex sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{store: make(map[interface{}]interface{})}
}

func (s *MemoryStore) Put(key, value interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.store[key] = value
	//fmt.Println(s.store)
	return nil
}

func (s *MemoryStore) Get(key interface{}) (interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.store[key]; ok {
		return v, nil
	}

	return nil, ErrNotExist
}

func (s *MemoryStore) Del(key interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.store, key)
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
	ClearBeans(tableName string)
}

type idNode struct {
	tbName    string
	id        int64
	lastVisit time.Time
}

type sqlNode struct {
	sql       string
	lastVisit time.Time
}

func newNode(tbName string, id int64) *idNode {
	return &idNode{tbName, id, time.Now()}
}

// LRUCacher implements Cacher according to LRU algorithm
type LRUCacher struct {
	idList   *list.List
	sqlList  *list.List
	idIndex  map[string]map[interface{}]*list.Element
	sqlIndex map[string]map[interface{}]*list.Element
	store    CacheStore
	Max      int
	mutex    sync.Mutex
	expired  int
}

func NewLRUCacher(store CacheStore, max int) *LRUCacher {
	cacher := &LRUCacher{store: store, idList: list.New(),
		sqlList: list.New(), Max: max}
	cacher.sqlIndex = make(map[string]map[interface{}]*list.Element)
	cacher.idIndex = make(map[string]map[interface{}]*list.Element)
	return cacher
}

func (m *LRUCacher) GetIds(tableName, sql string) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.sqlIndex[tableName]; !ok {
		m.sqlIndex[tableName] = make(map[interface{}]*list.Element)
	}
	if v, err := m.store.Get(sql); err == nil {
		if el, ok := m.sqlIndex[tableName][sql]; !ok {
			el = m.sqlList.PushBack(sql)
			m.sqlIndex[tableName][sql] = el
		} else {
			m.sqlList.MoveToBack(el)
		}
		return v
	} else {
		if el, ok := m.sqlIndex[tableName][sql]; ok {
			delete(m.sqlIndex[tableName], sql)
			m.sqlList.Remove(el)
		}
	}

	return nil
}

func (m *LRUCacher) GetBean(tableName string, id int64) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.idIndex[tableName]; !ok {
		m.idIndex[tableName] = make(map[interface{}]*list.Element)
	}
	if v, err := m.store.Get(genId(tableName, id)); err == nil {
		if el, ok := m.idIndex[tableName][id]; ok {
			m.idList.MoveToBack(el)
		} else {
			el = m.idList.PushBack(newNode(tableName, id))
			m.idIndex[tableName][id] = el
		}
		return v
	} else {
		// store bean is not exist, then remove memory's index
		if _, ok := m.idIndex[tableName][id]; ok {
			m.delBean(tableName, id)
			m.clearIds(tableName)
		}
		return nil
	}
}

func (m *LRUCacher) clearIds(tableName string) {
	//fmt.Println("clear ids")
	if tis, ok := m.sqlIndex[tableName]; ok {
		for sql, v := range tis {
			m.sqlList.Remove(v)
			m.store.Del(sql)
		}
	}
	m.sqlIndex[tableName] = make(map[interface{}]*list.Element)
}

func (m *LRUCacher) ClearIds(tableName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.clearIds(tableName)
}

func (m *LRUCacher) clearBeans(tableName string) {
	//fmt.Println("clear beans")
	if tis, ok := m.idIndex[tableName]; ok {
		//fmt.Println("before clear", len(m.idIndex[tableName]))
		for id, v := range tis {
			m.idList.Remove(v)
			tid := genId(tableName, id.(int64))
			m.store.Del(tid)
		}
		//fmt.Println("after clear", len(m.idIndex[tableName]))
	}
	m.idIndex[tableName] = make(map[interface{}]*list.Element)
}

func (m *LRUCacher) ClearBeans(tableName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.clearBeans(tableName)
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
	/*if m.sqlList.Len() > m.Max {
		e := m.sqlList.Front()
		node := e.Value.(*idNode)
		m.delBean(node.tbName, node.id)
	}*/
}

func (m *LRUCacher) PutBean(tableName string, id int64, obj interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var el *list.Element
	var ok bool

	if el, ok = m.idIndex[tableName][id]; !ok {
		el = m.idList.PushBack(newNode(tableName, id))
		m.idIndex[tableName][id] = el
	}

	m.store.Put(genId(tableName, id), obj)
	if m.idList.Len() > m.Max {
		e := m.idList.Front()
		node := e.Value.(*idNode)
		m.delBean(node.tbName, node.id)
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

func (m *LRUCacher) delBean(tableName string, id int64) {
	tid := genId(tableName, id)
	if el, ok := m.idIndex[tableName][tid]; ok {
		delete(m.idIndex[tableName], tid)
		m.idList.Remove(el)
		m.clearIds(tableName)
	}
	m.store.Del(tid)
}

func (m *LRUCacher) DelBean(tableName string, id int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.delBean(tableName, id)
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
