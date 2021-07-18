package rpc

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
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

func (s *Server) RegisterName(name string, method interface{}) error {
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
	srvc.replyType = srvc.methodType.In(1)
	if srvc.replyType.Kind() != reflect.Ptr {
		errMsg := fmt.Sprintf("param %v is not a pointer", srvc.replyType.Name())
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

func (s *Server) FindService(methodName string) (argv, replyv reflect.Value, argvNotPtr bool, srvc *service, err error) {
	v, ok := s.serviceMap.Load(methodName)
	if !ok {
		errMsg := fmt.Sprintf("method [%v] is not found", methodName)
		err = errors.New(errMsg)
		return
	}
	srvc = v.(*service)
	if srvc.argType.Kind() == reflect.Ptr {
		argv = reflect.New(srvc.argType.Elem())
	} else {
		argv = reflect.New(srvc.argType)
		argvNotPtr = true
	}
	replyv = reflect.New(srvc.replyType.Elem())

	return argv, replyv, argvNotPtr, srvc, nil
}

func (s *Server) collectError(codec *CodeC, wg *sync.WaitGroup) {
	for {
		err := <-s.errorC
		log.Println(err)
		if err == io.EOF {
			codec.closed = true
			wg.Wait()
			codec.conn.Close()
		}
	}
}

func (s *Server) Run(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(fmt.Sprintf("listen err: %v", err))
	}
	defer listener.Close()
	log.Println(fmt.Sprintf("bind: %s, start listening...", address))
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(fmt.Sprintf("accept err: %v", err))
		}
		bufWriter := bufio.NewWriter(conn)
		enc := gob.NewEncoder(bufWriter)
		dec := gob.NewDecoder(conn)
		codec := &CodeC{
			conn:   conn,
			dec:    dec,
			enc:    enc,
			encBuf: bufWriter,
		}
		var wg sync.WaitGroup
		go s.collectError(codec, &wg)
		go s.ServeLoop(codec, &wg)
	}
}
func (s *Server) writeError(enc *CodeC, res *Response, message string, writing *sync.Mutex) error {
	res.Error = message
	return s.writeResponse(enc, res, nil, writing)
}

func (s *Server) writeResponse(codec *CodeC, res *Response, replyv interface{}, writing *sync.Mutex) error {
	writing.Lock()
	defer writing.Unlock()
	if err := codec.enc.Encode(res); err != nil {
		return err
	}
	if replyv != nil {
		if err := codec.enc.Encode(replyv); err != nil {
			return err
		}
	}
	return codec.encBuf.Flush()
}

func (s *Server) callAndWriteRes(srvc *service, argv, replyv reflect.Value, res *Response, codec *CodeC, writing *sync.Mutex, wg *sync.WaitGroup) {
	returnValues := srvc.method.Call([]reflect.Value{argv, replyv})
	errValue := returnValues[0].Interface()
	if errValue != nil {
		writeErr := s.writeError(codec, res, errValue.(error).Error(), writing)
		if writeErr != nil {
			s.errorC <- writeErr
			wg.Done()
			return
		}
	}

	if writeErr := s.writeResponse(codec, res, replyv.Interface(), writing); writeErr != nil {
		s.errorC <- writeErr
		wg.Done()
		return
	}
	wg.Done()
}
func (s *Server) ServeLoop(codec *CodeC, wg *sync.WaitGroup) {
	var writing sync.Mutex

	for {
		if codec.closed {
			return
		}
		req := ReqPool.GetReq()
		res := ResPool.GetRes()
		// read req
		err := codec.dec.Decode(req)
		res.Seq = req.Seq
		if err != nil {
			s.errorC <- err
			writeErr := s.writeError(codec, res, err.Error(), &writing)
			if writeErr != nil {
				s.errorC <- writeErr
			}
			continue
		}
		argv, replyv, argNotPtr, srvc, err := s.FindService(req.ServiceName)
		if err != nil {
			writeErr := s.writeError(codec, res, err.Error(), &writing)
			if writeErr != nil {
				s.errorC <- writeErr
			}
			continue
		}
		// read arg
		err = codec.dec.Decode(argv.Interface())
		if err != nil {
			s.errorC <- err
			writeErr := s.writeError(codec, res, err.Error(), &writing)
			if writeErr != nil {
				s.errorC <- writeErr
			}
			continue
		}
		if argNotPtr {
			argv = argv.Elem()
		}

		wg.Add(1) // set done when the callAndWriteRes return
		go s.callAndWriteRes(srvc, argv, replyv, res, codec, &writing, wg)
	}
}

func main() {

}
