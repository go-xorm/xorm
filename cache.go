package xorm

import (
	"container/list"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// default cache expired time
	CacheExpired = 60 * time.Minute
	// not use now
	CacheMaxMemory = 256
	// evey ten minutes to clear all expired nodes
	CacheGcInterval = 10 * time.Minute
	// each time when gc to removed max nodes
	CacheGcMaxRemoved = 20
)

// CacheStore is a interface to store cache
type CacheStore interface {
	Put(key, value interface{}) error
	Get(key interface{}) (interface{}, error)
	Del(key interface{}) error
}

// MemoryStore implements CacheStore provide local machine
// memory store
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

// Cacher is an interface to provide cache
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
	tbName    string
	sql       string
	lastVisit time.Time
}

func newIdNode(tbName string, id int64) *idNode {
	return &idNode{tbName, id, time.Now()}
}

func newSqlNode(tbName, sql string) *sqlNode {
	return &sqlNode{tbName, sql, time.Now()}
}

// LRUCacher implements Cacher according to LRU algorithm
type LRUCacher struct {
	idList     *list.List
	sqlList    *list.List
	idIndex    map[string]map[interface{}]*list.Element
	sqlIndex   map[string]map[interface{}]*list.Element
	store      CacheStore
	Max        int
	mutex      sync.Mutex
	Expired    time.Duration
	maxSize    int
	GcInterval time.Duration
}

func newLRUCacher(store CacheStore, expired time.Duration, maxSize int, max int) *LRUCacher {
	cacher := &LRUCacher{store: store, idList: list.New(),
		sqlList: list.New(), Expired: expired, maxSize: maxSize,
		GcInterval: CacheGcInterval, Max: max,
		sqlIndex: make(map[string]map[interface{}]*list.Element),
		idIndex:  make(map[string]map[interface{}]*list.Element),
	}
	cacher.RunGC()
	return cacher
}

func NewLRUCacher(store CacheStore, max int) *LRUCacher {
	return newLRUCacher(store, CacheExpired, CacheMaxMemory, max)
}

func NewLRUCacher2(store CacheStore, expired time.Duration, max int) *LRUCacher {
	return newLRUCacher(store, expired, 0, max)
}

//func NewLRUCacher3(store CacheStore, expired time.Duration, maxSize int) *LRUCacher {
//	return newLRUCacher(store, expired, maxSize, 0)
//}

// RunGC run once every m.GcInterval
func (m *LRUCacher) RunGC() {
	time.AfterFunc(m.GcInterval, func() {
		m.RunGC()
		m.GC()
	})
}

// GC check ids lit and sql list to remove all element expired
func (m *LRUCacher) GC() {
	//fmt.Println("begin gc ...")
	//defer fmt.Println("end gc ...")
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var removedNum int
	for e := m.idList.Front(); e != nil; {
		if removedNum <= CacheGcMaxRemoved &&
			time.Now().Sub(e.Value.(*idNode).lastVisit) > m.Expired {
			removedNum++
			next := e.Next()
			//fmt.Println("removing ...", e.Value)
			node := e.Value.(*idNode)
			m.delBean(node.tbName, node.id)
			e = next
		} else {
			//fmt.Printf("removing %d cache nodes ..., left %d\n", removedNum, m.idList.Len())
			break
		}
	}

	removedNum = 0
	for e := m.sqlList.Front(); e != nil; {
		if removedNum <= CacheGcMaxRemoved &&
			time.Now().Sub(e.Value.(*sqlNode).lastVisit) > m.Expired {
			removedNum++
			next := e.Next()
			//fmt.Println("removing ...", e.Value)
			node := e.Value.(*sqlNode)
			m.delIds(node.tbName, node.sql)
			e = next
		} else {
			//fmt.Printf("removing %d cache nodes ..., left %d\n", removedNum, m.sqlList.Len())
			break
		}
	}
}

// Get all bean's ids according to sql and parameter from cache
func (m *LRUCacher) GetIds(tableName, sql string) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.sqlIndex[tableName]; !ok {
		m.sqlIndex[tableName] = make(map[interface{}]*list.Element)
	}
	if v, err := m.store.Get(sql); err == nil {
		if el, ok := m.sqlIndex[tableName][sql]; !ok {
			el = m.sqlList.PushBack(newSqlNode(tableName, sql))
			m.sqlIndex[tableName][sql] = el
		} else {
			lastTime := el.Value.(*sqlNode).lastVisit
			// if expired, remove the node and return nil
			if time.Now().Sub(lastTime) > m.Expired {
				m.delIds(tableName, sql)
				return nil
			}
			m.sqlList.MoveToBack(el)
			el.Value.(*sqlNode).lastVisit = time.Now()
		}
		return v
	} else {
		m.delIds(tableName, sql)
	}

	return nil
}

// Get bean according tableName and id from cache
func (m *LRUCacher) GetBean(tableName string, id int64) interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, ok := m.idIndex[tableName]; !ok {
		m.idIndex[tableName] = make(map[interface{}]*list.Element)
	}
	tid := genId(tableName, id)
	if v, err := m.store.Get(tid); err == nil {
		if el, ok := m.idIndex[tableName][id]; ok {
			lastTime := el.Value.(*idNode).lastVisit
			// if expired, remove the node and return nil
			if time.Now().Sub(lastTime) > m.Expired {
				m.delBean(tableName, id)
				//m.clearIds(tableName)
				return nil
			}
			m.idList.MoveToBack(el)
			el.Value.(*idNode).lastVisit = time.Now()
		} else {
			el = m.idList.PushBack(newIdNode(tableName, id))
			m.idIndex[tableName][id] = el
		}
		return v
	} else {
		// store bean is not exist, then remove memory's index
		m.delBean(tableName, id)
		//m.clearIds(tableName)
		return nil
	}
}

// Clear all sql-ids mapping on table tableName from cache
func (m *LRUCacher) clearIds(tableName string) {
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
	if tis, ok := m.idIndex[tableName]; ok {
		for id, v := range tis {
			m.idList.Remove(v)
			tid := genId(tableName, id.(int64))
			m.store.Del(tid)
		}
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
		el = m.sqlList.PushBack(newSqlNode(tableName, sql))
		m.sqlIndex[tableName][sql] = el
	} else {
		el.Value.(*sqlNode).lastVisit = time.Now()
	}
	m.store.Put(sql, ids)
	if m.sqlList.Len() > m.Max {
		e := m.sqlList.Front()
		node := e.Value.(*sqlNode)
		m.delIds(node.tbName, node.sql)
	}
}

func (m *LRUCacher) PutBean(tableName string, id int64, obj interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var el *list.Element
	var ok bool

	if el, ok = m.idIndex[tableName][id]; !ok {
		el = m.idList.PushBack(newIdNode(tableName, id))
		m.idIndex[tableName][id] = el
	} else {
		el.Value.(*idNode).lastVisit = time.Now()
	}

	m.store.Put(genId(tableName, id), obj)
	if m.idList.Len() > m.Max {
		e := m.idList.Front()
		node := e.Value.(*idNode)
		m.delBean(node.tbName, node.id)
	}
}

func (m *LRUCacher) delIds(tableName, sql string) {
	if _, ok := m.sqlIndex[tableName]; ok {
		if el, ok := m.sqlIndex[tableName][sql]; ok {
			delete(m.sqlIndex[tableName], sql)
			m.sqlList.Remove(el)
		}
	}
	m.store.Del(sql)
}

func (m *LRUCacher) DelIds(tableName, sql string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.delIds(tableName, sql)
}

func (m *LRUCacher) delBean(tableName string, id int64) {
	tid := genId(tableName, id)
	if el, ok := m.idIndex[tableName][id]; ok {
		delete(m.idIndex[tableName], id)
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
