package xorm

import (
	"errors"
	"strings"
	"time"
)

type mymysql struct {
	mysql
	proto   string
	raddr   string
	laddr   string
	timeout time.Duration
	db      string
	user    string
	passwd  string
}

func (db *mymysql) Init(drivername, uri string) error {
	db.mysql.base.init(drivername, uri)
	pd := strings.SplitN(uri, "*", 2)
	if len(pd) == 2 {
		// Parse protocol part of URI
		p := strings.SplitN(pd[0], ":", 2)
		if len(p) != 2 {
			return errors.New("Wrong protocol part of URI")
		}
		db.proto = p[0]
		options := strings.Split(p[1], ",")
		db.raddr = options[0]
		for _, o := range options[1:] {
			kv := strings.SplitN(o, "=", 2)
			var k, v string
			if len(kv) == 2 {
				k, v = kv[0], kv[1]
			} else {
				k, v = o, "true"
			}
			switch k {
			case "laddr":
				db.laddr = v
			case "timeout":
				to, err := time.ParseDuration(v)
				if err != nil {
					return err
				}
				db.timeout = to
			default:
				return errors.New("Unknown option: " + k)
			}
		}
		// Remove protocol part
		pd = pd[1:]
	}
	// Parse database part of URI
	dup := strings.SplitN(pd[0], "/", 3)
	if len(dup) != 3 {
		return errors.New("Wrong database part of URI")
	}
	db.dbname = dup[0]
	db.user = dup[1]
	db.passwd = dup[2]

	return nil
}
