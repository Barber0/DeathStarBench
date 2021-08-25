package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"unsafe"
)

func TestAlpha(t *testing.T) {
	resv := &Reservation{
		HotelId:      "123",
		CustomerName: "abcd",
	}
	myStr := "hahah"
	ptSz := unsafe.Sizeof(resv)
	objSz := unsafe.Sizeof(*resv)
	myInt := 123
	myInt32 := 123
	fmt.Println(
		ptSz,
		objSz,
		unsafe.Sizeof(myStr),
		unsafe.Sizeof(myInt),
		unsafe.Sizeof(myInt32),
	)
}

type BsonObj struct {
	Name string `bson:"name"`
	Age  int    `bson:"age"`
}

func TestBson(t *testing.T) {
	barr := []bson.M{
		{
			"name": "abc",
			"age":  123,
		},
		{
			"name": "def",
			"age":  456,
		},
	}

	bts, err := bson.Marshal(barr)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(len(bts))

	var oarr bson.D
	err = bson.Unmarshal(bts, &oarr)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(len(oarr))
}
