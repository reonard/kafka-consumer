package db

import (
	"fmt"
	"github.com/globalsign/mgo/bson"
	"testing"
)

func TestMongo(t *testing.T){

	InitDB("127.0.0.1:27017")
	//insertData()
	queryData()

}

func insertData(){

	s:=GetSession()
	c:=s.DB("pilot").C("device1")

	dt := bson.M{}
	dt["hehe"] = 1
	dt["test"] = 10

	err := c.Insert(dt)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("inserted")

}

func queryData(){

	s:=GetSession()
	defer s.Close()
	var a map[string]interface{}
	c:=s.DB("pilot").C("device1")
	c.Find(bson.M{"test":bson.M{"$gt":9}, "hehe":bson.M{"$let":2}}).One(&a)
	fmt.Println(a)
}