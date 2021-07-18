package rpc

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net"
	"sync"
)

type Caller interface {
	Call(serviceName string, arg interface{}, res interface{}) error
}

type Pending struct {
	doneC chan bool
	Reply interface{}
}

type Client struct {
	conn *net.TCPConn
	pendingMap sync.Map //map[int]Pending
	enc *gob.Encoder
	dec *gob.Decoder
	encBuf *bufio.Writer
	sending sync.Mutex
	reading sync.Mutex
	seq int
	next *Client
}

type ClientPool struct {
	nClients int
	Cursor *Client
	clientLock sync.Mutex
}

func (cp *ClientPool) GetClient() *Client {
	cp.clientLock.Lock()
	defer func() {
		cp.Cursor = cp.Cursor.next
		cp.clientLock.Unlock()
	}()
	return cp.Cursor
}

func (cp *ClientPool) addClient(c *Client) {
	cp.clientLock.Lock()
	cp.clientLock.Unlock()
	if cp.Cursor == nil {
		cp.Cursor = c
		cp.nClients = 1
		cp.Cursor.next = cp.Cursor
	} else {
		cp.Cursor.next,c.next = c, cp.Cursor.next
		cp.nClients += 1
	}
}

func (cp *ClientPool) deleteClient(c *Client) *Client {
	return nil
}

func (cp *ClientPool) Call(serviceName string, arg interface{}, res interface{}) error {
	var c Caller = cp.GetClient()
	return c.Call(serviceName, arg, res)
}

func DialMany(addressList []string) (Caller, error) {
	newPool := &ClientPool{}
	for _, address := range addressList {
		c, err := Dial(address)
		if err != nil {
			return nil, err
		}
		newPool.addClient(c)
	}
	return newPool, nil
}

func (c *Client) readRes() {
	for {
		res := ResPool.GetRes()
		err := c.dec.Decode(res)
		if err != nil {
			log.Println("decode response error:", err.Error())
			if err == io.EOF {
				break
			}
			continue
		}
		v, ok := c.pendingMap.LoadAndDelete(res.Seq)
		if !ok {
			log.Println("pending task with seq of", res.Seq, "is not found.")
			continue
		}
		err = c.dec.Decode(v.(*Pending).Reply)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		v.(*Pending).doneC <- true
		ResPool.FreeRes(res)
	}
}

func Dial(address string) (*Client, error) {
	c := &Client{}
	rAddr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil {
		return nil, err
	}
	c.conn, err = net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		return nil, err
	}
	bufWriter := bufio.NewWriter(c.conn)
	c.encBuf = bufWriter
	c.enc = gob.NewEncoder(bufWriter)
	c.dec = gob.NewDecoder(c.conn)
	go c.readRes()
	return c, nil
}

func (c *Client) call(req *Request, arg interface{}, pending *Pending) error {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.seq += 1
	req.Seq = c.seq
	if err := c.enc.Encode(req); err != nil {
		return err
	}
	if err := c.enc.Encode(arg); err != nil {
		return err
	}
	c.pendingMap.Store(req.Seq, pending)
	return c.encBuf.Flush()
}

func (c *Client) Call(serviceName string, arg interface{}, res interface{}) error {
	req := ReqPool.GetReq()
	defer ReqPool.FreeReq(req)
	req.ServiceName = serviceName
	pending := &Pending{
		doneC: make(chan bool),
		Reply: res,
	}
	err := c.call(req, arg, pending)
	if err != nil {
		return err
	}
	<-pending.doneC
	return nil
}