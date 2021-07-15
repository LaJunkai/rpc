package rpc

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"net"
)

type PTP struct {
	lBytes []byte
	Length uint64
	Content []byte
}

func NewPTP() *PTP {
	return &PTP{lBytes: make([]byte, 8)}
}


func WriteAsPTPBytes(conn net.Conn, v interface{}) (error) {
	bytesContent, err := json.Marshal(v)
	if err != nil {
		return err
	}
	lBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lBytes, uint64(len(bytesContent)))
	if i, err := conn.Write(lBytes); err != nil || i != 8 {
		return err
	}
	if i, err := conn.Write(bytesContent); err != nil || i != len(bytesContent) {
		return err
	}
	return nil
}

func ReadPTP(r *bufio.Reader) (*PTP, error) {
	newPTP := NewPTP()
	if _, err := r.Read(newPTP.lBytes); err != nil {
		return nil, err
	}
	newPTP.Length = binary.BigEndian.Uint64(newPTP.lBytes)
	newPTP.Content = make([]byte, newPTP.Length)
	if _, err := r.Read(newPTP.Content); err != nil {
		return nil, err
	}
	return newPTP, nil
}
