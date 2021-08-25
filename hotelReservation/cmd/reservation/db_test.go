package main

import (
	"fmt"
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
