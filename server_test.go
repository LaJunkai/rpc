package rpc

import "testing"

type A struct {

}

func (a *A) ExampleM()  {

}
func TestServer_Register(t *testing.T) {
	a := A{}
	server := Server{}
	err := server.Register(a.ExampleM, "")
	println(err.Error())
}
