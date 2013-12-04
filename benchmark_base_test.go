package xorm

import (
	"database/sql"
	"testing"
)

type BigStruct struct {
	Id       int64
	Name     string
	Title    string
	Age      string
	Alias    string
	NickName string
}

func doBenchDriverInsert(engine *Engine, db *sql.DB, b *testing.B) {
	b.StopTimer()
	err := engine.CreateTables(&BigStruct{})
	if err != nil {
		b.Error(err)
		return
	}

	doBenchDriverInsertS(db, b)

	err = engine.DropTables(&BigStruct{})
	if err != nil {
		b.Error(err)
		return
	}
}

func doBenchDriverInsertS(db *sql.DB, b *testing.B) {
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

func doBenchDriverFind(engine *Engine, db *sql.DB, b *testing.B) {
	b.StopTimer()
	err := engine.CreateTables(&BigStruct{})
	if err != nil {
		b.Error(err)
		return
	}

	doBenchDriverFindS(db, b)

	err = engine.DropTables(&BigStruct{})
	if err != nil {
		b.Error(err)
		return
	}
}

func doBenchDriverFindS(db *sql.DB, b *testing.B) {
	b.StopTimer()
	for i := 0; i < 100; i++ {
		_, err := db.Exec(`insert into big_struct (name, title, age, alias, nick_name) 
			values ('fafdasf', 'fadfa', 'afadfsaf', 'fadfafdsafd', 'fadfafdsaf')`)
		if err != nil {
			b.Error(err)
			return
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Query("select * from big_struct limit 50")
		if err != nil {
			b.Error(err)
			return
		}
	}
	b.StopTimer()
}

func doBenchInsert(engine *Engine, b *testing.B) {
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

func doBenchFind(engine *Engine, b *testing.B) {
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
	for i := 0; i < b.N; i++ {
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

func doBenchFindPtr(engine *Engine, b *testing.B) {
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
	for i := 0; i < b.N; i++ {
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
