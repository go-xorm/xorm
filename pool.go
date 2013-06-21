package xorm

import (
	"database/sql"
	"fmt"
	//"sync"
	"sync/atomic"
	//"time"
)

type IConnectionPool interface {
	RetrieveDB(engine *Engine) (*sql.DB, error)
	ReleaseDB(engine *Engine, db *sql.DB)
}

type NoneConnectPool struct {
}

var ConnectionNum int32 = 0

func (p NoneConnectPool) RetrieveDB(engine *Engine) (db *sql.DB, err error) {
	atomic.AddInt32(&ConnectionNum, 1)
	db, err = engine.OpenDB()
	fmt.Printf("--open a connection--%x\n", &db)
	return
}

func (p NoneConnectPool) ReleaseDB(engine *Engine, db *sql.DB) {
	atomic.AddInt32(&ConnectionNum, -1)
	fmt.Printf("--close a connection--%x\n", &db)
	db.Close()
}

/*type SimpleConnectPool struct {
	releasedSessions []*sql.DB
	cur              int
	usingSessions    map[*sql.DB]time.Time
	maxWaitTimeOut   int
	mutex            *sync.Mutex
}

func (p SimpleConnectPool) RetrieveDB(engine *Engine) (*sql.DB, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	var db *sql.DB = nil
	var err error = nil
	fmt.Printf("%x, rbegin - released:%v, using:%v\n", &p, p.cur+1, len(p.usingSessions))
	if p.cur < 0 {
		ConnectionNum = ConnectionNum + 1
		fmt.Printf("new %v\n", ConnectionNum)
		db, err = engine.OpenDB()
		if err != nil {
			return nil, err
		}
		p.usingSessions[db] = time.Now()
	} else {
		db = p.releasedSessions[p.cur]
		p.usingSessions[db] = time.Now()
		p.releasedSessions[p.cur] = nil
		p.cur = p.cur - 1
		fmt.Println("release one")
	}

	fmt.Printf("%x, rend - released:%v, using:%v\n", &p, p.cur+1, len(p.usingSessions))
	return db, nil
}

func (p SimpleConnectPool) ReleaseDB(engine *Engine, db *sql.DB) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	fmt.Printf("%x, lbegin - released:%v, using:%v\n", &p, p.cur+1, len(p.usingSessions))
	if p.cur >= 29 {
		ConnectionNum = ConnectionNum - 1
		db.Close()
	} else {
		p.cur = p.cur + 1
		p.releasedSessions[p.cur] = db
	}
	delete(p.usingSessions, db)
	fmt.Printf("%x, lend - released:%v, using:%v\n", &p, p.cur+1, len(p.usingSessions))
}*/
