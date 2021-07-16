package rpc

import (
	"testing"
)

func EM(a int, b int) error {
	return nil
}

func B(a int, b *int) error {
	*b = a + 1
	return nil
}

func TestServer_Register(t *testing.T) {
	server := NewServer()
	err := server.Register("111", B)
	Assert(err != nil && err.Error() == "param int is not a pointer")
}
func TestServer_Run(t *testing.T) {
	server := NewServer()
	server.Register("111", B)
	server.Run(":9999")
}

