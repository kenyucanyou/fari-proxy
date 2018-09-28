package server

import (
	"encoding/binary"
	"github.com/fari-proxy/encryption"
	"github.com/fari-proxy/service"
	"log"
	"net"
)

type server struct {
	*service.Service
}

func NewServer(addr, password string) *server {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	c := encryption.NewCipher([]byte(password))
	return &server{
		&service.Service{
			Cipher:     c,
			ListenAddr: tcpAddr,
		},
	}
}

func (s *server) Listen() {
	listen, err := net.ListenTCP("tcp", s.ListenAddr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("启动成功,监听在 %s:%d, 密码: %s", s.ListenAddr.IP, s.ListenAddr.Port, s.Cipher.Password)
	defer listen.Close()

	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			log.Fatalf("%s", err.Error())
			continue
		}
		conn.SetLinger(0)
		go s.handle(conn)
	}
}

func (s *server) handle(conn *net.TCPConn) {
	defer conn.Close()
	/*
		RFC 1928 - IETF
		https://www.ietf.org/rfc/rfc1928.txt
	*/

	// Establish socks5 connection
	// Step one: receive client request [version, nmethods, methods]
	buf := make([]byte, 256)
	_, err := s.Decode(conn, buf)
	if err != nil || (buf[0] != 0x05) {
		return
	}
	// Step two: send to client 0x05,0x00 [version, method]
	s.Encode(conn, []byte{0x05, 0x00})

	// Step three: get the command and destination server address
	n, err := s.Decode(conn, buf)
	if err != nil {
		return
	}

	if buf[1] != 0x01 { // Only support connect
		return
	}
	// Parse destination addr and port
	var desIP []byte
	switch buf[3] {
	case 0x01:
		desIP = buf[4 : 4+net.IPv4len]
	case 0x03:
		ipAddr, err := net.ResolveIPAddr("ip", string(buf[5:n-2]))
		if err != nil {
			return
		}
		desIP = ipAddr.IP
	case 0x04:
		desIP = buf[4 : 4+net.IPv6len]
	default:
		return
	}
	dstPort := buf[n-2 : n]
	dstAddr := &net.TCPAddr{
		IP:   desIP,
		Port: int(binary.BigEndian.Uint16(dstPort)),
	}
	// Step four: connect to the destination server and send a reply to client
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		return
	} else {
		defer dstServer.Close()
		dstServer.SetLinger(0)
		s.Encode(conn, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	log.Printf("Connect to destination addr %s", dstAddr.String())
	// Read data from the peer-end to destination server
	go func() {
		err := s.DecodeTransfer(dstServer, conn)
		if err != nil {
			conn.Close()
			dstServer.Close()
		}
	}()
	// Read data from destination server to the peer-end
	s.EncodeTransfer(conn, dstServer)
}
