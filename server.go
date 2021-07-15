package rpc

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"sync"
)

type Server struct {
	serviceMap sync.Map // map[string]*service
	errorC     chan error
}

func NewServer() *Server {
	return &Server{errorC: make(chan error, 8)}
}

func (s *Server) Register(name string, method interface{}) error {
	srvc := &service{}
	srvc.method = reflect.ValueOf(method)
	srvc.methodType = reflect.TypeOf(method)
	srvc.name = name
	// method
	if n := srvc.methodType.NumIn(); n != 2 {
		errMsg := fmt.Sprintf("2 args are excepted instead of %v", n)
		return errors.New(errMsg)
	}
	srvc.argType = srvc.methodType.In(0)
	srvc.resType = srvc.methodType.In(1)
	if srvc.resType.Kind() != reflect.Ptr {
		errMsg := fmt.Sprintf("param %v is not a pointer", srvc.resType.Name())
		return errors.New(errMsg)
	}
	if srvc.name == "" {
		errMsg := fmt.Sprintf("no name for services %v", srvc.methodType.String())
		return errors.New(errMsg)
	}
	if _, ok := s.serviceMap.LoadOrStore(srvc.name, srvc); ok {
		errMsg := fmt.Sprintf("registered method [%v] may be overwritted.\n", srvc.name)
		return errors.New(errMsg)
	}
	return nil
}

func (s *Server) FindService(methodName string) (argv, resv reflect.Value, srvc *service, err error) {
	v, ok := s.serviceMap.Load(methodName)
	if !ok {
		errMsg := fmt.Sprintf("method [%v] is not found", methodName)
		err = errors.New(errMsg)
		return
	}
	srvc = v.(*service)
	argv = reflect.New(srvc.argType.Elem())
	resv = reflect.New(srvc.resType.Elem())
	returnValues := srvc.method.Call([]reflect.Value{argv, resv})
	errInter := returnValues[0].Interface()
	if errInter != nil {
		err = errInter.(error)
		return
	}
	return argv, resv, srvc, nil
}

func (s *Server) collectError() {
	for {
		err := <-s.errorC
		log.Fatal(err)
	}
}

func (s *Server) ListenAndServe(address string) {
	// 绑定监听地址
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(fmt.Sprintf("listen err: %v", err))
	}
	defer listener.Close()
	log.Println(fmt.Sprintf("bind: %s, start listening...", address))
	go s.collectError()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(fmt.Sprintf("accept err: %v", err))
		}
		bufWriter := bufio.NewWriter(conn)
		enc := gob.NewEncoder(bufWriter)
		dec := gob.NewDecoder(conn)
		go s.ServeLoop(dec, enc)
	}
}
func (s *Server) writeError(enc *gob.Encoder, res *Response, message string, writing *sync.Mutex) error {
	res.error = message
	return s.writeResponse(enc, res, writing)
}

func (s *Server) writeResponse(enc *gob.Encoder, res *Response, writing *sync.Mutex) error {
	writing.Lock()
	defer writing.Unlock()
	if err := enc.Encode(res); err != nil {
		return err
	}
	return nil
}

func (s *Server) callAndWriteRes(srvc *service, argv, resv reflect.Value, res *Response, enc *gob.Encoder, writing *sync.Mutex) {
	returnValues := srvc.method.Call([]reflect.Value{argv, resv})
	errValue := returnValues[0].Interface()
	if errValue != nil {
		writeErr := s.writeError(enc, res, errValue.(error).Error(), writing)
		if writeErr != nil {
			s.errorC <- writeErr
			return
		}
	}
	if writeErr := s.writeResponse(enc, res, writing); writeErr != nil {
		s.errorC <- writeErr
		return
	}
}
func (s *Server) ServeLoop(dec *gob.Decoder, enc *gob.Encoder) {
	// 使用 bufio 标准库提供的缓冲区功能
	var writing sync.Mutex
	for {
		req := ReqPool.GetReq()
		res := ResPool.GetRes()
		res.seq = req.seq
		dec.Decode(req)
		argv, resv, srvc, err := s.FindService(req.ServiceName)
		if err != nil {
			writeErr := s.writeError(enc, res, err.Error(), &writing)
			if writeErr != nil {
				s.errorC <- writeErr
			}
			continue
		}
		go s.callAndWriteRes(srvc, argv, resv, res, enc, &writing)
	}
}

func main() {

}
