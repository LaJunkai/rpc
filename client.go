package rpc

import (
	"bufio"
	"encoding/gob"
	"log"
	"net"
	"sync"
)

type Pending struct {
	doneC chan bool
	res interface{}
}

type Client struct {
	conn *net.TCPConn
	pendingMap sync.Map //map[int]Pending
	enc *gob.Encoder
	dec *gob.Decoder
	sending sync.Mutex
	reading sync.Mutex
	seq int
}

func (c *Client) readRes() {
	for {
		res := ResPool.GetRes()
		err := c.dec.Decode(res)
		if err != nil {
			log.Fatal(err.Error())
			continue
		}
		v, ok := c.pendingMap.LoadAndDelete(res.seq)
		if !ok {
			continue
		}
		v.(chan bool) <- true
		ResPool.FreeRes(res)
	}
}

func (c *Client) Dial(address string) error {
	rAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return err
	}
	c.conn, err = net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		return err
	}
	c.enc = gob.NewEncoder(c.conn)
	c.dec = gob.NewDecoder(bufio.NewReader(c.conn))
	go c.readRes()
	return nil
}

func (c *Client) call(req *Request, arg interface{}, pending *Pending) error {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.seq += 1
	req.seq = c.seq
	if err := c.enc.Encode(req); err != nil {
		return err
	}
	if err := c.enc.Encode(arg); err != nil {
		return err
	}
	c.pendingMap.Store(req.seq, pending)
	return nil
}

func (c *Client) Call(serviceName string, arg interface{}, res interface{}) error {
	req := ReqPool.GetReq()
	defer ReqPool.FreeReq(req)
	req.ServiceName = serviceName
	pending := &Pending{
		doneC: make(chan bool),
		res:   ResPool,
	}
	err := c.call(req, arg, pending)
	if err != nil {
		return err
	}
	<-pending.doneC
	return nil
}