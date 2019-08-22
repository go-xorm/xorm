package xorm

import (
	"fmt"
	"testing"
)

type Contact struct {
	UserId      int    `json:"user_id"`
	Name        string `json:"name" jorm:"real_name" xorm:"real_name"`
	Age         int    `json:"age"`
	PhoneNumber string `json:"phone_number"`
	HomeAddress string `json:"home_address"`
	CreateTime  string `json:"create_time"`
}

func TestEngine_CallProcedure(t *testing.T) {

	mysqlClient, err := NewEngine("mysql", "jerry:Ming521.@tcp(jerry.igoogle.ink:3306)/db_test?charset=utf8")
	if err != nil {
		fmt.Println("err:", err)
	}
	mysqlClient.ShowSQL(true)

	//test Get
	contact := new(Contact)
	_, err = mysqlClient.StartProcedure("query_contact", 1, 6).InParams("付明明").Get(contact)
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println("contact:", contact)

	//test Find
	contactList := make([]Contact, 0)
	err = mysqlClient.StartProcedure("query_contact", 1, 6).InParams("付明明").Find(&contactList)
	if err != nil {
		fmt.Println("err:", err)
	}
	fmt.Println("contactList:", contactList)

}
