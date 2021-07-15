package rpc

import "testing"

func EM(a int, b int) error {
	return nil
}

func B(a int, b *int) error {
	*b = a + *b
	return nil
}

func TestServer_Register(t *testing.T) {
	server := Server{}
	err := server.Register("111", EM)
	if err != nil {
		panic(err)
	}
}
func TestServer_Register2(t *testing.T) {

}
