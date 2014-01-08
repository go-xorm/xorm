package drivers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lunny/xorm/core"
)

func init() {
	core.RegisterDriver("postgres", &pqDriver{})
}

type pqDriver struct {
}

type values map[string]string

func (vs values) Set(k, v string) {
	vs[k] = v
}

func (vs values) Get(k string) (v string) {
	return vs[k]
}

func errorf(s string, args ...interface{}) {
	panic(fmt.Errorf("pq: %s", fmt.Sprintf(s, args...)))
}

func parseOpts(name string, o values) {
	if len(name) == 0 {
		return
	}

	name = strings.TrimSpace(name)

	ps := strings.Split(name, " ")
	for _, p := range ps {
		kv := strings.Split(p, "=")
		if len(kv) < 2 {
			errorf("invalid option: %q", p)
		}
		o.Set(kv[0], kv[1])
	}
}

func (p *pqDriver) Parse(driverName, dataSourceName string) (*core.Uri, error) {
	db := &core.Uri{DbType: core.POSTGRES}
	o := make(values)
	parseOpts(dataSourceName, o)

	db.DbName = o.Get("dbname")
	if db.DbName == "" {
		return nil, errors.New("dbname is empty")
	}
	return db, nil
}
