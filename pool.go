package xorm

import (
	"database/sql"
	//"fmt"
	//"sync"
	//"time"
)

type IConnectionPool interface {
	RetrieveDB(engine *Engine) (*sql.DB, error)
	ReleaseDB(engine *Engine, db *sql.DB)
}

type NoneConnectPool struct {
}

func (p NoneConnectPool) RetrieveDB(engine *Engine) (db *sql.DB, err error) {
	db, err = engine.OpenDB()
	return
}

func (p NoneConnectPool) ReleaseDB(engine *Engine, db *sql.DB) {
	db.Close()
}

/*
var (
	total int = 0
)

type SimpleConnectPool struct {
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
		total = total + 1
		fmt.Printf("new %v\n", total)
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
		db.Close()
	} else {
		p.cur = p.cur + 1
		p.releasedSessions[p.cur] = db
	}
	delete(p.usingSessions, db)
	fmt.Printf("%x, lend - released:%v, using:%v\n", &p, p.cur+1, len(p.usingSessions))
}*/
