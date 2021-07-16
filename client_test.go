package rpc

import (
	"fmt"
	"testing"
)

func TestClient_Dial(t *testing.T) {
	c, err := Dial("127.0.0.1:9999")
	if err != nil {
		println(err.Error())
	}
	println(c)
	Assert(err == nil)
}

func TestClient_Call(t *testing.T) {
	c, err := Dial("127.0.0.1:9999")
	if err != nil {
		println(err.Error())
	}
	Assert(err == nil)
	a, b := 1, 1
	err = c.Call("111", a, &b)
	if err != nil {
		println(err.Error())
	}
	fmt.Println("a =", a, "b =", b)
}
