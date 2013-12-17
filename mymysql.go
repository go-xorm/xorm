package xorm

import (
    "errors"
    "strings"
    "time"
)

type mymysql struct {
    mysql
}

type mymysqlParser struct {
}

func (p *mymysqlParser) parse(driverName, dataSourceName string) (*uri, error) {
    db := &uri{dbType: MYSQL}

    pd := strings.SplitN(dataSourceName, "*", 2)
    if len(pd) == 2 {
        // Parse protocol part of URI
        p := strings.SplitN(pd[0], ":", 2)
        if len(p) != 2 {
            return nil, errors.New("Wrong protocol part of URI")
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
                    return nil, err
                }
                db.timeout = to
            default:
                return nil, errors.New("Unknown option: " + k)
            }
        }
        // Remove protocol part
        pd = pd[1:]
    }
    // Parse database part of URI
    dup := strings.SplitN(pd[0], "/", 3)
    if len(dup) != 3 {
        return nil, errors.New("Wrong database part of URI")
    }
    db.dbName = dup[0]
    db.user = dup[1]
    db.passwd = dup[2]

    return db, nil
}

func (db *mymysql) Init(drivername, uri string) error {
    return db.mysql.base.init(&mymysqlParser{}, drivername, uri)
}
