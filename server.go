package rpc

import (
	"bufio"
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
}

func (s *Server) Register(method interface{}, name string) error {
	service := &service{}
	service.callable = reflect.ValueOf(method)
	service.typ = reflect.TypeOf(method)
	if name == "" {
		// if no name passed by user
		service.name = reflect.Indirect(service.callable).Type().Name()
	} else {
		service.name = name
	}
	fmt.Println(reflect.Indirect(service.callable).)
	if service.name == "" {
		return errors.New(fmt.Sprintf("no name for services %v", service.typ.String()))
	}
	if _, ok := s.serviceMap.LoadOrStore(service.name, service); !ok {
		return errors.New(fmt.Sprintf("registered method [%v] may be overwritted.\n", service.name))
	}
	return nil
}
func (s *Server)ListenAndServe(address string) {
	// 绑定监听地址
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(fmt.Sprintf("listen err: %v", err))
	}
	defer listener.Close()
	log.Println(fmt.Sprintf("bind: %s, start listening...", address))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(fmt.Sprintf("accept err: %v", err))
		}
		go s.ServeLoop(conn)
	}
}

func (s *Server)ServeLoop(conn net.Conn) {
	// 使用 bufio 标准库提供的缓冲区功能
	reader := bufio.NewReader(conn)
	for {
		ptp, err := ReadPTP(reader)
		if err != nil {
			if err == io.EOF {
				log.Println("connection closed")
				return
			} else {
				log.Println(err.Error(), err)
				return
			}
		}
		println(ptp)
	}
}

func main() {

}