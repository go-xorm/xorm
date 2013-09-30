package xorm

import (
	"database/sql"
	//"fmt"
	"sync"
	//"sync/atomic"
	"container/list"
	"time"
)

// Interface IConnecPool is a connection pool interface, all implements should implement
// Init, RetrieveDB, ReleaseDB and Close methods.
// Init for init when engine be created or invoke SetPool
// RetrieveDB for requesting a connection to db;
// ReleaseDB for releasing a db connection;
// Close for invoking when engine.Close
type IConnectPool interface {
	Init(engine *Engine) error
	RetrieveDB(engine *Engine) (*sql.DB, error)
	ReleaseDB(engine *Engine, db *sql.DB)
	Close(engine *Engine) error
	SetMaxIdleConns(conns int)
	MaxIdleConns() int
	SetMaxConns(conns int)
	MaxConns() int
}

// Struct NoneConnectPool is a implement for IConnectPool. It provides directly invoke driver's
// open and release connection function
type NoneConnectPool struct {
}

// NewNoneConnectPool new a NoneConnectPool.
func NewNoneConnectPool() IConnectPool {
	return &NoneConnectPool{}
}

// Init do nothing
func (p *NoneConnectPool) Init(engine *Engine) error {
	return nil
}

// RetrieveDB directly open a connection
func (p *NoneConnectPool) RetrieveDB(engine *Engine) (db *sql.DB, err error) {
	db, err = engine.OpenDB()
	return
}

// ReleaseDB directly close a connection
func (p *NoneConnectPool) ReleaseDB(engine *Engine, db *sql.DB) {
	db.Close()
}

// Close do nothing
func (p *NoneConnectPool) Close(engine *Engine) error {
	return nil
}

func (p *NoneConnectPool) SetMaxIdleConns(conns int) {
}

func (p *NoneConnectPool) MaxIdleConns() int {
	return 0
}

// not implemented
func (p *NoneConnectPool) SetMaxConns(conns int) {
}

// not implemented
func (p *NoneConnectPool) MaxConns() int {
	return -1
}

// Struct SysConnectPool is a simple wrapper for using system default connection pool.
// About the system connection pool, you can review the code database/sql/sql.go
// It's currently default Pool implments.
type SysConnectPool struct {
	db           *sql.DB
	maxIdleConns int
	maxConns     int
	curConns     int
	mutex        *sync.Mutex
	queue        *list.List
}

// NewSysConnectPool new a SysConnectPool.
func NewSysConnectPool() IConnectPool {
	return &SysConnectPool{}
}

// Init create a db immediately and keep it util engine closed.
func (s *SysConnectPool) Init(engine *Engine) error {
	db, err := engine.OpenDB()
	if err != nil {
		return err
	}
	s.db = db
	s.maxIdleConns = 2
	s.maxConns = -1
	s.curConns = 0
	s.mutex = &sync.Mutex{}
	s.queue = list.New()
	return nil
}

type node struct {
	mutex sync.Mutex
	cond  *sync.Cond
}

func newCondNode() *node {
	n := &node{}
	n.cond = sync.NewCond(&n.mutex)
	return n
}

// RetrieveDB just return the only db
func (s *SysConnectPool) RetrieveDB(engine *Engine) (db *sql.DB, err error) {
	/*if s.maxConns > 0 {
		fmt.Println("before retrieve")
		s.mutex.Lock()
		for s.curConns >= s.maxConns {
			fmt.Println("before waiting...", s.curConns, s.queue.Len())
			s.mutex.Unlock()
			n := NewNode()
			n.cond.L.Lock()
			s.queue.PushBack(n)
			n.cond.Wait()
			n.cond.L.Unlock()
			s.mutex.Lock()
			fmt.Println("after waiting...", s.curConns, s.queue.Len())
		}
		s.curConns += 1
		s.mutex.Unlock()
		fmt.Println("after retrieve")
	}*/
	return s.db, nil
}

// ReleaseDB do nothing
func (s *SysConnectPool) ReleaseDB(engine *Engine, db *sql.DB) {
	/*if s.maxConns > 0 {
		s.mutex.Lock()
		fmt.Println("before release", s.queue.Len())
		s.curConns -= 1

		if e := s.queue.Front(); e != nil {
			n := e.Value.(*node)
			//n.cond.L.Lock()
			n.cond.Signal()
			fmt.Println("signaled...")
			s.queue.Remove(e)
			//n.cond.L.Unlock()
		}
		fmt.Println("after released", s.queue.Len())
		s.mutex.Unlock()
	}*/
}

// Close closed the only db
func (p *SysConnectPool) Close(engine *Engine) error {
	return p.db.Close()
}

func (p *SysConnectPool) SetMaxIdleConns(conns int) {
	p.db.SetMaxIdleConns(conns)
	p.maxIdleConns = conns
}

func (p *SysConnectPool) MaxIdleConns() int {
	return p.maxIdleConns
}

// not implemented
func (p *SysConnectPool) SetMaxConns(conns int) {
	p.maxConns = conns
	//p.db.SetMaxOpenConns(conns)
}

// not implemented
func (p *SysConnectPool) MaxConns() int {
	return p.maxConns
}

// NewSimpleConnectPool new a SimpleConnectPool
func NewSimpleConnectPool() IConnectPool {
	return &SimpleConnectPool{releasedConnects: make([]*sql.DB, 10),
		usingConnects:  map[*sql.DB]time.Time{},
		cur:            -1,
		maxWaitTimeOut: 14400,
		maxIdleConns:   10,
		mutex:          &sync.Mutex{},
	}
}

// Struct SimpleConnectPool is a simple implementation for IConnectPool.
// It's a custom connection pool and not use system connection pool.
// Opening or Closing a database connection must be enter a lock.
// This implements will be improved in furture.
type SimpleConnectPool struct {
	releasedConnects []*sql.DB
	cur              int
	usingConnects    map[*sql.DB]time.Time
	maxWaitTimeOut   int
	mutex            *sync.Mutex
	maxIdleConns     int
}

func (s *SimpleConnectPool) Init(engine *Engine) error {
	return nil
}

// RetrieveDB get a connection from connection pool
func (p *SimpleConnectPool) RetrieveDB(engine *Engine) (*sql.DB, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	var db *sql.DB = nil
	var err error = nil
	//fmt.Printf("%x, rbegin - released:%v, using:%v\n", &p, p.cur+1, len(p.usingConnects))
	if p.cur < 0 {
		db, err = engine.OpenDB()
		if err != nil {
			return nil, err
		}
		p.usingConnects[db] = time.Now()
	} else {
		db = p.releasedConnects[p.cur]
		p.usingConnects[db] = time.Now()
		p.releasedConnects[p.cur] = nil
		p.cur = p.cur - 1
	}

	//fmt.Printf("%x, rend - released:%v, using:%v\n", &p, p.cur+1, len(p.usingConnects))
	return db, nil
}

// ReleaseDB release a db from connection pool
func (p *SimpleConnectPool) ReleaseDB(engine *Engine, db *sql.DB) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	//fmt.Printf("%x, lbegin - released:%v, using:%v\n", &p, p.cur+1, len(p.usingConnects))
	if p.cur >= p.maxIdleConns-1 {
		db.Close()
	} else {
		p.cur = p.cur + 1
		p.releasedConnects[p.cur] = db
	}
	delete(p.usingConnects, db)
	//fmt.Printf("%x, lend - released:%v, using:%v\n", &p, p.cur+1, len(p.usingConnects))
}

// Close release all db
func (p *SimpleConnectPool) Close(engine *Engine) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for len(p.releasedConnects) > 0 {
		p.releasedConnects[0].Close()
		p.releasedConnects = p.releasedConnects[1:]
	}

	return nil
}

func (p *SimpleConnectPool) SetMaxIdleConns(conns int) {
	p.maxIdleConns = conns
}

func (p *SimpleConnectPool) MaxIdleConns() int {
	return p.maxIdleConns
}

// not implemented
func (p *SimpleConnectPool) SetMaxConns(conns int) {
}

// not implemented
func (p *SimpleConnectPool) MaxConns() int {
	return -1
}
