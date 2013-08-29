package xorm

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

/*
CREATE TABLE `userinfo` (
	`id` INT(10) NULL AUTO_INCREMENT,
	`username` VARCHAR(64) NULL,
	`departname` VARCHAR(64) NULL,
	`created` DATE NULL,
	PRIMARY KEY (`uid`)
);
CREATE TABLE `userdeatail` (
	`id` INT(10) NULL,
	`intro` TEXT NULL,
	`profile` TEXT NULL,
	PRIMARY KEY (`uid`)
);
*/

type Userinfo struct {
	Uid        int64  `xorm:"id pk not null autoincr"`
	Username   string `xorm:"unique"`
	Departname string
	Alias      string `xorm:"-"`
	Created    time.Time
	Detail     Userdetail `xorm:"detail_id int(11)"`
	Height     float64
	Avatar     []byte
	IsMan      bool
}

type Userdetail struct {
	Id      int64
	Intro   string `xorm:"text"`
	Profile string `xorm:"varchar(2000)"`
}

func directCreateTable(engine *Engine, t *testing.T) {
	err := engine.DropTables(&Userinfo{}, &Userdetail{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.CreateTables(&Userinfo{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func mapper(engine *Engine, t *testing.T) {
	err := engine.UnMap(&Userinfo{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.Map(&Userinfo{}, &Userdetail{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.DropAll()
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.CreateAll()
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func insert(engine *Engine, t *testing.T) {
	user := Userinfo{0, "xiaolunwen", "dev", "lunny", time.Now(),
		Userdetail{Id: 1}, 1.78, []byte{1, 2, 3}, true}
	_, err := engine.Insert(&user)
	fmt.Println(user.Uid)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func query(engine *Engine, t *testing.T) {
	sql := "select * from userinfo"
	results, err := engine.Query(sql)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(results)
}

func exec(engine *Engine, t *testing.T) {
	sql := "update userinfo set username=? where id=?"
	res, err := engine.Exec(sql, "xiaolun", 1)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(res)
}

func insertAutoIncr(engine *Engine, t *testing.T) {
	// auto increment insert
	user := Userinfo{Username: "xiaolunwen", Departname: "dev", Alias: "lunny", Created: time.Now(),
		Detail: Userdetail{Id: 1}, Height: 1.78, Avatar: []byte{1, 2, 3}, IsMan: true}
	_, err := engine.Insert(&user)
	fmt.Println(user.Uid)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func insertMulti(engine *Engine, t *testing.T) {
	//engine.InsertMany = true
	users := []Userinfo{
		{Username: "xlw", Departname: "dev", Alias: "lunny2", Created: time.Now()},
		{Username: "xlw2", Departname: "dev", Alias: "lunny3", Created: time.Now()},
		{Username: "xlw11", Departname: "dev", Alias: "lunny2", Created: time.Now()},
		{Username: "xlw22", Departname: "dev", Alias: "lunny3", Created: time.Now()},
	}
	_, err := engine.Insert(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func insertTwoTable(engine *Engine, t *testing.T) {
	userdetail := Userdetail{Id: 1, Intro: "I'm a very beautiful women.", Profile: "sfsaf"}
	userinfo := Userinfo{Username: "xlw3", Departname: "dev", Alias: "lunny4", Created: time.Now(), Detail: userdetail}

	_, err := engine.Insert(&userinfo, &userdetail)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

type Condi map[string]interface{}

func update(engine *Engine, t *testing.T) {
	// update by id
	user := Userinfo{Username: "xxx", Height: 1.2}
	_, err := engine.Id(1).Update(&user)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	condi := Condi{"username": "zzz", "height": 0.0, "departname": ""}
	_, err = engine.Table(&user).Id(1).Update(condi)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	_, err = engine.Update(&Userinfo{Username: "yyy"}, &user)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func testdelete(engine *Engine, t *testing.T) {
	user := Userinfo{Uid: 1}
	_, err := engine.Delete(&user)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func get(engine *Engine, t *testing.T) {
	user := Userinfo{Uid: 2}

	has, err := engine.Get(&user)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if has {
		fmt.Println(user)
	} else {
		fmt.Println("no record id is 2")
	}
}

func cascadeGet(engine *Engine, t *testing.T) {
	user := Userinfo{Uid: 11}

	has, err := engine.Get(&user)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if has {
		fmt.Println(user)
	} else {
		fmt.Println("no record id is 2")
	}
}

func find(engine *Engine, t *testing.T) {
	users := make([]Userinfo, 0)

	err := engine.Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func findMap(engine *Engine, t *testing.T) {
	users := make(map[int64]Userinfo)

	err := engine.Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func count(engine *Engine, t *testing.T) {
	user := Userinfo{Departname: "dev"}
	total, err := engine.Count(&user)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Printf("Total %d records!!!", total)
}

func where(engine *Engine, t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.Where("id > ?", 2).Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func in(engine *Engine, t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.In("id", 1, 2, 3).Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)

	ids := []interface{}{1, 2, 3}
	err = engine.Where("id > ?", 2).In("id", ids...).Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func limit(engine *Engine, t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.Limit(2, 1).Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func order(engine *Engine, t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.OrderBy("id desc").Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func join(engine *Engine, t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.Join("LEFT", "userdetail", "userinfo.id=userdetail.id").Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func having(engine *Engine, t *testing.T) {
	users := make([]Userinfo, 0)
	err := engine.GroupBy("username").Having("username='xlw'").Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(users)
}

func transaction(engine *Engine, t *testing.T) {
	counter := func() {
		total, err := engine.Count(&Userinfo{})
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("----now total %v records\n", total)
	}

	counter()
	defer counter()
	session := engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		t.Error(err)
		panic(err)
	}
	//session.IsAutoRollback = false
	user1 := Userinfo{Username: "xiaoxiao", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
	}
	user2 := Userinfo{Username: "yyy"}
	_, err = session.Where("uid = ?", 0).Update(&user2)
	if err != nil {
		session.Rollback()
		fmt.Println(err)
		//t.Error(err)
		return
	}

	_, err = session.Delete(&user2)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
	}

	err = session.Commit()
	if err != nil {
		t.Error(err)
		panic(err)
	}
	panic(err)
}

func combineTransaction(engine *Engine, t *testing.T) {
	counter := func() {
		total, err := engine.Count(&Userinfo{})
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("----now total %v records\n", total)
	}

	counter()
	defer counter()
	session := engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		t.Error(err)
		panic(err)
	}
	//session.IsAutoRollback = false
	user1 := Userinfo{Username: "xiaoxiao2", Departname: "dev", Alias: "lunny", Created: time.Now()}
	_, err = session.Insert(&user1)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
	}
	user2 := Userinfo{Username: "zzz"}
	_, err = session.Where("id = ?", 0).Update(&user2)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
	}

	_, err = session.Exec("delete from userinfo where username = ?", user2.Username)
	if err != nil {
		session.Rollback()
		t.Error(err)
		panic(err)
	}

	err = session.Commit()
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func table(engine *Engine, t *testing.T) {
	err := engine.DropTables("user_user")
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.Table("user_user").CreateTable(&Userinfo{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func createMultiTables(engine *Engine, t *testing.T) {
	session := engine.NewSession()
	defer session.Close()

	user := &Userinfo{}
	err := session.Begin()
	if err != nil {
		t.Error(err)
		panic(err)
	}
	for i := 0; i < 10; i++ {
		tableName := fmt.Sprintf("user_%v", i)

		err = session.DropTable(tableName)
		if err != nil {
			session.Rollback()
			t.Error(err)
			panic(err)
		}

		err = session.Table(tableName).CreateTable(user)
		if err != nil {
			session.Rollback()
			t.Error(err)
			panic(err)
		}
	}
	err = session.Commit()
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func tableOp(engine *Engine, t *testing.T) {
	user := Userinfo{Username: "tablexiao", Departname: "dev", Alias: "lunny", Created: time.Now()}
	tableName := fmt.Sprintf("user_%v", len(user.Username))
	id, err := engine.Table(tableName).Insert(&user)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	_, err = engine.Table(tableName).Get(&Userinfo{Username: "tablexiao"})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	users := make([]Userinfo, 0)
	err = engine.Table(tableName).Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	_, err = engine.Table(tableName).Id(id).Update(&Userinfo{Username: "tableda"})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	_, err = engine.Table(tableName).Id(id).Delete(&Userinfo{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func testCharst(engine *Engine, t *testing.T) {
	err := engine.DropTables("user_charset")
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.Charset("utf8").Table("user_charset").CreateTable(&Userinfo{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

func testStoreEngine(engine *Engine, t *testing.T) {
	err := engine.DropTables("user_store_engine")
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.StoreEngine("InnoDB").Table("user_store_engine").CreateTable(&Userinfo{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

type tempUser struct {
	Id       int64
	Username string
}

func testCols(engine *Engine, t *testing.T) {
	users := []Userinfo{}
	err := engine.Cols("id, username").Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	fmt.Println(users)

	tmpUsers := []tempUser{}
	err = engine.Table("userinfo").Cols("id, username").Find(&tmpUsers)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(tmpUsers)
}

type tempUser2 struct {
	tempUser   `xorm:"extends"`
	Departname string
}

func testExtends(engine *Engine, t *testing.T) {
	err := engine.DropTables(&tempUser2{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.CreateTables(&tempUser2{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu := &tempUser2{tempUser{0, "extends"}, "dev depart"}
	_, err = engine.Insert(tu)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu2 := &tempUser2{}
	_, err = engine.Get(tu2)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu3 := &tempUser2{tempUser{0, "extends update"}, ""}
	_, err = engine.Id(tu2.Id).Update(tu3)
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

type allCols struct {
	Bit       int   `xorm:"BIT"`
	TinyInt   int8  `xorm:"TINYINT"`
	SmallInt  int16 `xorm:"SMALLINT"`
	MediumInt int32 `xorm:"MEDIUMINT"`
	Int       int   `xorm:"INT"`
	Integer   int   `xorm:"INTEGER"`
	BigInt    int64 `xorm:"BIGINT"`

	Char       string `xorm:"CHAR(12)"`
	Varchar    string `xorm:"VARCHAR(54)"`
	TinyText   string `xorm:"TINYTEXT"`
	Text       string `xorm:"TEXT"`
	MediumText string `xorm:"MEDIUMTEXT"`
	LongText   string `xorm:"LONGTEXT"`
	Binary     string `xorm:"BINARY"`
	VarBinary  string `xorm:"VARBINARY(12)"`

	Date      time.Time `xorm:"DATE"`
	DateTime  time.Time `xorm:"DATETIME"`
	Time      time.Time `xorm:"TIME"`
	TimeStamp time.Time `xorm:"TIMESTAMP"`

	Decimal float64 `xorm:"DECIMAL"`
	Numeric float64 `xorm:"NUMERIC"`

	Real   float32 `xorm:"REAL"`
	Float  float32 `xorm:"FLOAT"`
	Double float64 `xorm:"DOUBLE"`

	TinyBlob   []byte `xorm:"TINYBLOB"`
	Blob       []byte `xorm:"BLOB"`
	MediumBlob []byte `xorm:"MEDIUMBLOB"`
	LongBlob   []byte `xorm:"LONGBLOB"`
	Bytea      []byte `xorm:"BYTEA"`

	Bool bool `xorm:"BOOL"`

	Serial int `xorm:"SERIAL"`
	//BigSerial int64 `xorm:"BIGSERIAL"`
}

func testColTypes(engine *Engine, t *testing.T) {
	err := engine.DropTables(&allCols{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = engine.CreateTables(&allCols{})
	if err != nil {
		t.Error(err)
		panic(err)
	}
}

type MyInt int
type MyUInt uint
type MyFloat float64
type MyString string

func (s MyString) FromDB(data []byte) error {
	s = MyString(string(data))
	return nil
}

func (s MyString) ToDB() ([]byte, error) {
	return []byte(string(s)), nil
}

type MyStruct struct {
	Type MyInt
	U    MyUInt
	F    MyFloat
	//S    MyString
	//IA []MyInt
	//UA        []MyUInt
	//FA        []MyFloat
	//SA        []MyString
	//NameArray []string
	Name string
	//UIA       []uint
	UI uint
}

func testCustomType(engine *Engine, t *testing.T) {
	err := engine.CreateTables(&MyStruct{})
	i := MyStruct{Name: "Test", Type: MyInt(1)}
	_, err = engine.Insert(&i)
	if err != nil {
		t.Error(err)
		panic(err)
		return
	}

	has, err := engine.Get(&i)
	if err != nil {
		t.Error(err)
		panic(err)
	} else if !has {
		t.Error(errors.New("should get one record"))
		panic(err)
	}
}

func testTrans(engine *Engine, t *testing.T) {
}

func testAll(engine *Engine, t *testing.T) {
	directCreateTable(engine, t)
	mapper(engine, t)
	insert(engine, t)
	query(engine, t)
	exec(engine, t)
	insertAutoIncr(engine, t)
	insertMulti(engine, t)
	insertTwoTable(engine, t)
	update(engine, t)
	testdelete(engine, t)
	get(engine, t)
	cascadeGet(engine, t)
	find(engine, t)
	findMap(engine, t)
	count(engine, t)
	where(engine, t)
	in(engine, t)
	limit(engine, t)
	order(engine, t)
	join(engine, t)
	having(engine, t)
	transaction(engine, t)
	combineTransaction(engine, t)
	table(engine, t)
	createMultiTables(engine, t)
	tableOp(engine, t)
	testCols(engine, t)
	testCharst(engine, t)
	testStoreEngine(engine, t)
	testExtends(engine, t)
	testColTypes(engine, t)
	testCustomType(engine, t)
}
