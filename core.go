package rpc

import "reflect"

var ReqPool RequestPool
var ResPool ResponsePool

func init() {
	ReqPool = &BaseRequestPool{}
	ResPool = &BaseResponsePool{}
}

type Request struct {
	seq int
	ServiceName string
	next *Request
}

type Response struct {
	seq int
	ServiceName string
	next *Request
	error string
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
	name     string
	method   reflect.Method
	callable reflect.Value
	typ      reflect.Type
	argTyp   reflect.Type
	resTyp   reflect.Type
}