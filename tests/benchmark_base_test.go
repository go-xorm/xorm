package xorm

import (
	"database/sql"
	"testing"

	"github.com/lunny/xorm"
)

type BigStruct struct {
	Id       int64
	Name     string
	Title    string
	Age      string
	Alias    string
	NickName string
}

func doBenchDriverInsert(db *sql.DB, b *testing.B) {
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Exec(`insert into big_struct (name, title, age, alias, nick_name) 
            values ('fafdasf', 'fadfa', 'afadfsaf', 'fadfafdsafd', 'fadfafdsaf')`)
		if err != nil {
			b.Error(err)
			return
		}
	}
	b.StopTimer()
}

func doBenchDriverFind(db *sql.DB, b *testing.B) {
	b.StopTimer()
	for i := 0; i < 50; i++ {
		_, err := db.Exec(`insert into big_struct (name, title, age, alias, nick_name) 
            values ('fafdasf', 'fadfa', 'afadfsaf', 'fadfafdsafd', 'fadfafdsaf')`)
		if err != nil {
			b.Error(err)
			return
		}
	}

	b.StartTimer()
	for i := 0; i < b.N/50; i++ {
		rows, err := db.Query("select * from big_struct limit 50")
		if err != nil {
			b.Error(err)
			return
		}
		for rows.Next() {
			s := &BigStruct{}
			rows.Scan(&s.Id, &s.Name, &s.Title, &s.Age, &s.Alias, &s.NickName)
		}
	}
	b.StopTimer()
}

func doBenchDriver(newdriver func() (*sql.DB, error), createTableSql,
	dropTableSql string, opFunc func(*sql.DB, *testing.B), t *testing.B) {
	db, err := newdriver()
	if err != nil {
		t.Error(err)
		return
	}
	defer db.Close()

	_, err = db.Exec(createTableSql)
	if err != nil {
		t.Error(err)
		return
	}

	opFunc(db, t)

	_, err = db.Exec(dropTableSql)
	if err != nil {
		t.Error(err)
		return
	}
}

func doBenchInsert(engine *xorm.Engine, b *testing.B) {
	b.StopTimer()
	bs := &BigStruct{0, "fafdasf", "fadfa", "afadfsaf", "fadfafdsafd", "fadfafdsaf"}
	err := engine.CreateTables(bs)
	if err != nil {
		b.Error(err)
		return
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		bs.Id = 0
		_, err = engine.Insert(bs)
		if err != nil {
			b.Error(err)
			return
		}
	}
	b.StopTimer()
	err = engine.DropTables(bs)
	if err != nil {
		b.Error(err)
		return
	}
}

func doBenchFind(engine *xorm.Engine, b *testing.B) {
	b.StopTimer()
	bs := &BigStruct{0, "fafdasf", "fadfa", "afadfsaf", "fadfafdsafd", "fadfafdsaf"}
	err := engine.CreateTables(bs)
	if err != nil {
		b.Error(err)
		return
	}

	for i := 0; i < 100; i++ {
		bs.Id = 0
		_, err = engine.Insert(bs)
		if err != nil {
			b.Error(err)
			return
		}
	}

	b.StartTimer()
	for i := 0; i < b.N/50; i++ {
		bss := new([]BigStruct)
		err = engine.Limit(50).Find(bss)
		if err != nil {
			b.Error(err)
			return
		}
	}
	b.StopTimer()
	err = engine.DropTables(bs)
	if err != nil {
		b.Error(err)
		return
	}
}

func doBenchFindPtr(engine *xorm.Engine, b *testing.B) {
	b.StopTimer()
	bs := &BigStruct{0, "fafdasf", "fadfa", "afadfsaf", "fadfafdsafd", "fadfafdsaf"}
	err := engine.CreateTables(bs)
	if err != nil {
		b.Error(err)
		return
	}

	for i := 0; i < 100; i++ {
		bs.Id = 0
		_, err = engine.Insert(bs)
		if err != nil {
			b.Error(err)
			return
		}
	}

	b.StartTimer()
	for i := 0; i < b.N/50; i++ {
		bss := new([]*BigStruct)
		err = engine.Limit(50).Find(bss)
		if err != nil {
			b.Error(err)
			return
		}
	}
	b.StopTimer()
	err = engine.DropTables(bs)
	if err != nil {
		b.Error(err)
		return
	}
}
