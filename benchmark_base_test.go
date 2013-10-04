package xorm

import (
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

func doBenchCacheFind(engine *Engine, b *testing.B) {
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
