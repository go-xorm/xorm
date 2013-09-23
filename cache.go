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
	Get(id interface{}) interface{}
	Put(id, obj interface{})
	Del(id interface{})
}

// LRUCacher implements Cacher according to LRU algorithm
type LRUCacher struct {
	name  string
	list  *list.List
	index map[interface{}]*list.Element
	store CacheStore
	Max   int
	mutex sync.RWMutex
}

func NewLRUCacher(store CacheStore, max int) *LRUCacher {
	return &LRUCacher{store: store, list: list.New(),
		index: make(map[interface{}]*list.Element), Max: max}
}

func (m *LRUCacher) Get(id interface{}) interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if v, err := m.store.Get(id); err == nil {
		el := m.index[id]
		m.list.MoveToBack(el)
		return v
	}
	return nil
}

func (m *LRUCacher) Put(id interface{}, obj interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	el := m.list.PushBack(id)
	m.index[id] = el
	m.store.Put(id, obj)
	if m.list.Len() > m.Max {
		e := m.list.Front()
		m.store.Del(e.Value)
		delete(m.index, e.Value)
		m.list.Remove(e)
	}
}

func (m *LRUCacher) Del(id interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if el, ok := m.index[id]; ok {
		m.store.Del(id)
		delete(m.index, el.Value)
		m.list.Remove(el)
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

func getCacheSql(m Cacher, sql string, args interface{}) ([]int64, error) {
	bytes := m.Get(genSqlKey(sql, args))
	if bytes == nil {
		return nil, errors.New("Not Exist")
	}
	objs := decodeIds(bytes.(string))
	return objs, nil
}

func putCacheSql(m Cacher, ids []int64, sql string, args interface{}) error {
	bytes := encodeIds(ids)
	m.Put(genSqlKey(sql, args), bytes)
	return nil
}

func delCacheSql(m Cacher, sql string, args interface{}) error {
	m.Del(genSqlKey(sql, args))
	return nil
}

func genSqlKey(sql string, args interface{}) string {
	return fmt.Sprintf("%v-%v", sql, args)
}

func genId(prefix string, id int64) string {
	return fmt.Sprintf("%v-%v", prefix, id)
}

func getCacheId(m Cacher, prefix string, id int64) interface{} {
	return m.Get(genId(prefix, id))
}

func putCacheId(m Cacher, prefix string, id int64, bean interface{}) error {
	m.Put(genId(prefix, id), bean)
	return nil
}

func delCacheId(m Cacher, prefix string, id int64) error {
	m.Del(genId(prefix, id))
	//TODO: should delete id from select
	return nil
}
