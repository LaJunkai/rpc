package rpc

import (
	"testing"
)

func EM(a int, b int) error {
	return nil
}

type A struct {

}
func (st *A) B(a int, b *int) error {
	*b = a + 1
	return nil
}
var st = A{}
func TestServer_Register(t *testing.T) {
	server := NewServer()
	err := server.RegisterName("111", st.B)
	Assert(err != nil && err.Error() == "param int is not a pointer")
}
func TestServer_Run(t *testing.T) {
	server := NewServer()
	server.RegisterName("111", st.B)
	server.Run(":9999")
}

