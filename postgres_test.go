package xorm

import (
	_ "github.com/bylevel/pq"
	"testing"
)

func TestPostgres(t *testing.T) {
	engine, err := NewEngine("postgres", "dbname=xorm_test sslmode=disable")
	defer engine.Close()
	if err != nil {
		t.Error(err)
		return
	}
	engine.ShowSQL = true

	testAll(engine, t)
}
