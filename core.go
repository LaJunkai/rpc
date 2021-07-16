package rpc

import (
	"bufio"
	"encoding/gob"
	"io"
	"reflect"
)

var ReqPool RequestPool
var ResPool ResponsePool

func init() {
	ReqPool = &BaseRequestPool{}
	ResPool = &BaseResponsePool{}
}

type Request struct {
	Seq         int
	ServiceName string
	next        *Request
}

type Response struct {
	Seq         int
	ServiceName string
	next        *Request
	Error       string
}

type RequestPool interface {
	GetReq() *Request
	FreeReq(req *Request)
}
type ResponsePool interface {
	GetRes() *Response
	FreeRes(res *Response)
}

type BaseRequestPool struct {

}

func (b *BaseRequestPool) GetReq() *Request {
	return &Request{}
}

func (b *BaseRequestPool) FreeReq(req *Request) {

}

type BaseResponsePool struct {

}

func (b *BaseResponsePool) GetRes() *Response {
	return &Response{}
}

func (b *BaseResponsePool) FreeRes(res *Response) {

}

type service struct {
	name       string
	method     reflect.Value
	methodType reflect.Type
	argType    reflect.Type
	replyType  reflect.Type
}

type CodeC struct {
	conn    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func Assert(b bool) {
	if !b {
		panic("the assert is not true")
	}
}