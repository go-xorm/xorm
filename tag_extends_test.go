// Copyright 2017 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xorm

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-xorm/core"
	"github.com/stretchr/testify/assert"
)

type tempUser struct {
	Id       int64
	Username string
}

type tempUser2 struct {
	TempUser   tempUser `xorm:"extends"`
	Departname string
}

type tempUser3 struct {
	Temp       *tempUser `xorm:"extends"`
	Departname string
}

type tempUser4 struct {
	TempUser2 tempUser2 `xorm:"extends"`
}

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

type UserAndDetail struct {
	Userinfo   `xorm:"extends"`
	Userdetail `xorm:"extends"`
}

func TestExtends(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&tempUser2{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&tempUser2{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu := &tempUser2{tempUser{0, "extends"}, "dev depart"}
	_, err = testEngine.Insert(tu)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu2 := &tempUser2{}
	_, err = testEngine.Get(tu2)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu3 := &tempUser2{tempUser{0, "extends update"}, ""}
	_, err = testEngine.ID(tu2.TempUser.Id).Update(tu3)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.DropTables(&tempUser4{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&tempUser4{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu8 := &tempUser4{tempUser2{tempUser{0, "extends"}, "dev depart"}}
	_, err = testEngine.Insert(tu8)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu9 := &tempUser4{}
	_, err = testEngine.Get(tu9)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if tu9.TempUser2.TempUser.Username != tu8.TempUser2.TempUser.Username || tu9.TempUser2.Departname != tu8.TempUser2.Departname {
		err = errors.New(fmt.Sprintln("not equal for", tu8, tu9))
		t.Error(err)
		panic(err)
	}

	tu10 := &tempUser4{tempUser2{tempUser{0, "extends update"}, ""}}
	_, err = testEngine.ID(tu9.TempUser2.TempUser.Id).Update(tu10)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.DropTables(&tempUser3{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&tempUser3{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu4 := &tempUser3{&tempUser{0, "extends"}, "dev depart"}
	_, err = testEngine.Insert(tu4)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	tu5 := &tempUser3{}
	_, err = testEngine.Get(tu5)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if tu5.Temp == nil {
		err = errors.New("error get data extends")
		t.Error(err)
		panic(err)
	}
	if tu5.Temp.Id != 1 || tu5.Temp.Username != "extends" ||
		tu5.Departname != "dev depart" {
		err = errors.New("error get data extends")
		t.Error(err)
		panic(err)
	}

	tu6 := &tempUser3{&tempUser{0, "extends update"}, ""}
	_, err = testEngine.ID(tu5.Temp.Id).Update(tu6)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	users := make([]tempUser3, 0)
	err = testEngine.Find(&users)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if len(users) != 1 {
		err = errors.New("error get data not 1")
		t.Error(err)
		panic(err)
	}

	assertSync(t, new(Userinfo), new(Userdetail))

	detail := Userdetail{
		Intro: "I'm in China",
	}
	_, err = testEngine.Insert(&detail)
	assert.NoError(t, err)

	_, err = testEngine.Insert(&Userinfo{
		Username: "lunny",
		Detail:   detail,
	})
	assert.NoError(t, err)

	var info UserAndDetail
	qt := testEngine.Quote
	ui := testEngine.TableName(new(Userinfo), true)
	ud := testEngine.TableName(&detail, true)
	uiid := testEngine.GetColumnMapper().Obj2Table("Id")
	udid := "detail_id"
	sql := fmt.Sprintf("select * from %s, %s where %s.%s = %s.%s",
		qt(ui), qt(ud), qt(ui), qt(udid), qt(ud), qt(uiid))
	b, err := testEngine.SQL(sql).NoCascade().Get(&info)
	assert.NoError(t, err)
	if !b {
		err = errors.New("should has lest one record")
		t.Error(err)
		panic(err)
	}
	fmt.Println(info)
	if info.Userinfo.Uid == 0 || info.Userdetail.Id == 0 {
		err = errors.New("all of the id should has value")
		t.Error(err)
		panic(err)
	}

	fmt.Println("----join--info2")
	var info2 UserAndDetail
	b, err = testEngine.Table(&Userinfo{}).
		Join("LEFT", qt(ud), qt(ui)+"."+qt("detail_id")+" = "+qt(ud)+"."+qt(uiid)).
		NoCascade().Get(&info2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	if !b {
		err = errors.New("should has lest one record")
		t.Error(err)
		panic(err)
	}
	if info2.Userinfo.Uid == 0 || info2.Userdetail.Id == 0 {
		err = errors.New("all of the id should has value")
		t.Error(err)
		panic(err)
	}
	fmt.Println(info2)

	fmt.Println("----join--infos2")
	var infos2 = make([]UserAndDetail, 0)
	err = testEngine.Table(&Userinfo{}).
		Join("LEFT", qt(ud), qt(ui)+"."+qt("detail_id")+" = "+qt(ud)+"."+qt(uiid)).
		NoCascade().
		Find(&infos2)
	if err != nil {
		t.Error(err)
		panic(err)
	}
	fmt.Println(infos2)
}

type MessageBase struct {
	Id     int64 `xorm:"int(11) pk autoincr"`
	TypeId int64 `xorm:"int(11) notnull"`
}

type Message struct {
	MessageBase `xorm:"extends"`
	Title       string    `xorm:"varchar(100) notnull"`
	Content     string    `xorm:"text notnull"`
	Uid         int64     `xorm:"int(11) notnull"`
	ToUid       int64     `xorm:"int(11) notnull"`
	CreateTime  time.Time `xorm:"datetime notnull created"`
}

type MessageUser struct {
	Id   int64
	Name string
}

type MessageType struct {
	Id   int64
	Name string
}

type MessageExtend3 struct {
	Message  `xorm:"extends"`
	Sender   MessageUser `xorm:"extends"`
	Receiver MessageUser `xorm:"extends"`
	Type     MessageType `xorm:"extends"`
}

type MessageExtend4 struct {
	Message     `xorm:"extends"`
	MessageUser `xorm:"extends"`
	MessageType `xorm:"extends"`
}

func TestExtends2(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&Message{}, &MessageUser{}, &MessageType{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&Message{}, &MessageUser{}, &MessageType{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	var sender = MessageUser{Name: "sender"}
	var receiver = MessageUser{Name: "receiver"}
	var msgtype = MessageType{Name: "type"}
	_, err = testEngine.Insert(&sender, &receiver, &msgtype)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	msg := Message{
		MessageBase: MessageBase{
			Id: msgtype.Id,
		},
		Title:   "test",
		Content: "test",
		Uid:     sender.Id,
		ToUid:   receiver.Id,
	}
	if testEngine.Dialect().DBType() == core.MSSQL {
		_, err = testEngine.Exec("SET IDENTITY_INSERT message ON")
		assert.NoError(t, err)
	}

	_, err = testEngine.Insert(&msg)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	var mapper = testEngine.GetTableMapper().Obj2Table
	var quote = testEngine.Quote
	userTableName := quote(testEngine.TableName(mapper("MessageUser"), true))
	typeTableName := quote(testEngine.TableName(mapper("MessageType"), true))
	msgTableName := quote(testEngine.TableName(mapper("Message"), true))

	list := make([]Message, 0)
	err = testEngine.Table(msgTableName).Join("LEFT", []string{userTableName, "sender"}, "`sender`.`"+mapper("Id")+"`="+msgTableName+".`"+mapper("Uid")+"`").
		Join("LEFT", []string{userTableName, "receiver"}, "`receiver`.`"+mapper("Id")+"`="+msgTableName+".`"+mapper("ToUid")+"`").
		Join("LEFT", []string{typeTableName, "type"}, "`type`.`"+mapper("Id")+"`="+msgTableName+".`"+mapper("Id")+"`").
		Find(&list)
	assert.NoError(t, err)

	if len(list) != 1 {
		err = errors.New(fmt.Sprintln("should have 1 message, got", len(list)))
		t.Error(err)
		panic(err)
	}

	if list[0].Id != msg.Id {
		err = errors.New(fmt.Sprintln("should message equal", list[0], msg))
		t.Error(err)
		panic(err)
	}
}

func TestExtends3(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&Message{}, &MessageUser{}, &MessageType{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&Message{}, &MessageUser{}, &MessageType{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	var sender = MessageUser{Name: "sender"}
	var receiver = MessageUser{Name: "receiver"}
	var msgtype = MessageType{Name: "type"}
	_, err = testEngine.Insert(&sender, &receiver, &msgtype)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	msg := Message{
		MessageBase: MessageBase{
			Id: msgtype.Id,
		},
		Title:   "test",
		Content: "test",
		Uid:     sender.Id,
		ToUid:   receiver.Id,
	}
	if testEngine.Dialect().DBType() == core.MSSQL {
		_, err = testEngine.Exec("SET IDENTITY_INSERT message ON")
		assert.NoError(t, err)
	}
	_, err = testEngine.Insert(&msg)
	assert.NoError(t, err)

	var mapper = testEngine.GetTableMapper().Obj2Table
	var quote = testEngine.Quote
	userTableName := quote(testEngine.TableName(mapper("MessageUser"), true))
	typeTableName := quote(testEngine.TableName(mapper("MessageType"), true))
	msgTableName := quote(testEngine.TableName(mapper("Message"), true))

	list := make([]MessageExtend3, 0)
	err = testEngine.Table(msgTableName).Join("LEFT", []string{userTableName, "sender"}, "`sender`.`"+mapper("Id")+"`="+msgTableName+".`"+mapper("Uid")+"`").
		Join("LEFT", []string{userTableName, "receiver"}, "`receiver`.`"+mapper("Id")+"`="+msgTableName+".`"+mapper("ToUid")+"`").
		Join("LEFT", []string{typeTableName, "type"}, "`type`.`"+mapper("Id")+"`="+msgTableName+".`"+mapper("Id")+"`").
		Find(&list)
	assert.NoError(t, err)

	if len(list) != 1 {
		err = errors.New(fmt.Sprintln("should have 1 message, got", len(list)))
		t.Error(err)
		panic(err)
	}

	if list[0].Message.Id != msg.Id {
		err = errors.New(fmt.Sprintln("should message equal", list[0].Message, msg))
		t.Error(err)
		panic(err)
	}

	if list[0].Sender.Id != sender.Id || list[0].Sender.Name != sender.Name {
		err = errors.New(fmt.Sprintln("should sender equal", list[0].Sender, sender))
		t.Error(err)
		panic(err)
	}

	if list[0].Receiver.Id != receiver.Id || list[0].Receiver.Name != receiver.Name {
		err = errors.New(fmt.Sprintln("should receiver equal", list[0].Receiver, receiver))
		t.Error(err)
		panic(err)
	}

	if list[0].Type.Id != msgtype.Id || list[0].Type.Name != msgtype.Name {
		err = errors.New(fmt.Sprintln("should msgtype equal", list[0].Type, msgtype))
		t.Error(err)
		panic(err)
	}
}

func TestExtends4(t *testing.T) {
	assert.NoError(t, prepareEngine())

	err := testEngine.DropTables(&Message{}, &MessageUser{}, &MessageType{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	err = testEngine.CreateTables(&Message{}, &MessageUser{}, &MessageType{})
	if err != nil {
		t.Error(err)
		panic(err)
	}

	var sender = MessageUser{Name: "sender"}
	var msgtype = MessageType{Name: "type"}
	_, err = testEngine.Insert(&sender, &msgtype)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	msg := Message{
		MessageBase: MessageBase{
			Id: msgtype.Id,
		},
		Title:   "test",
		Content: "test",
		Uid:     sender.Id,
	}
	if testEngine.Dialect().DBType() == core.MSSQL {
		_, err = testEngine.Exec("SET IDENTITY_INSERT message ON")
		assert.NoError(t, err)
	}
	_, err = testEngine.Insert(&msg)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	var mapper = testEngine.GetTableMapper().Obj2Table
	var quote = testEngine.Quote
	userTableName := quote(testEngine.TableName(mapper("MessageUser"), true))
	typeTableName := quote(testEngine.TableName(mapper("MessageType"), true))
	msgTableName := quote(testEngine.TableName(mapper("Message"), true))

	list := make([]MessageExtend4, 0)
	err = testEngine.Table(msgTableName).Join("LEFT", userTableName, userTableName+".`"+mapper("Id")+"`="+msgTableName+".`"+mapper("Uid")+"`").
		Join("LEFT", typeTableName, typeTableName+".`"+mapper("Id")+"`="+msgTableName+".`"+mapper("Id")+"`").
		Find(&list)
	if err != nil {
		t.Error(err)
		panic(err)
	}

	if len(list) != 1 {
		err = errors.New(fmt.Sprintln("should have 1 message, got", len(list)))
		t.Error(err)
		panic(err)
	}

	if list[0].Message.Id != msg.Id {
		err = errors.New(fmt.Sprintln("should message equal", list[0].Message, msg))
		t.Error(err)
		panic(err)
	}

	if list[0].MessageUser.Id != sender.Id || list[0].MessageUser.Name != sender.Name {
		err = errors.New(fmt.Sprintln("should sender equal", list[0].MessageUser, sender))
		t.Error(err)
		panic(err)
	}

	if list[0].MessageType.Id != msgtype.Id || list[0].MessageType.Name != msgtype.Name {
		err = errors.New(fmt.Sprintln("should msgtype equal", list[0].MessageType, msgtype))
		t.Error(err)
		panic(err)
	}
}
