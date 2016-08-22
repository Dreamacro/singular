package singular

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"

	pb "github.com/Dreamacro/singular/protobuf"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

// Client define a client
type Client struct {
	Sessions
	Name       string
	LocalAddr  string
	ServerAddr string
	tls        bool
	tlsConfig  *tls.Config
}

// NewClient return a client
func NewClient(name, localAddr, serverAddr string, tls bool, tlsConfig *tls.Config) *Client {
	return &Client{
		Name:       name,
		LocalAddr:  localAddr,
		ServerAddr: serverAddr,
		tls:        tls,
		tlsConfig:  tlsConfig,
	}
}

// Connect connect server
func (client *Client) Connect() error {
	var conn net.Conn
	var err error
	if client.tls {
		conn, err = tls.Dial("tcp", client.ServerAddr, client.tlsConfig)
	} else {
		conn, err = net.Dial("tcp", client.ServerAddr)
	}
	if err != nil {
		return err
	}

	con := NewConn(conn)

	req := &pb.Request{
		Meta:    pb.Request_NewProxy,
		Payload: client.Name,
	}
	buf, _ := proto.Marshal(req)
	con.Send(buf)

	buf, err = con.Receive()
	if err != nil {
		log.Errorf("receive buf: %v", err)
	}

	res := &pb.Request{}
	err = proto.Unmarshal(buf, res)

	if res.Meta != pb.Request_Assign {
		log.WithFields(log.Fields{"err": err, "data": res}).Info("Assign Response Error")
	}

	conn.Close()

	host, _, err := net.SplitHostPort(client.ServerAddr)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Info("SplitHostPort Error")
		return err
	}
	addr := fmt.Sprintf("%s:%s", host, res.Payload)
	client.Process(addr)
	return io.EOF
}

// Process working
func (client *Client) Process(addr string) {
	var conn net.Conn
	var err error
	if client.tls {
		conn, err = tls.Dial("tcp", addr, client.tlsConfig)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		log.Errorf("fail: %v", err)
	}
	log.Printf("%s Process: %s", client.Name, addr)
	client.Sessions = NewSessions(NewConn(conn))

	defer client.Close()

	for {
		buf, err := client.Sessions.Client.Receive()
		if err != nil {
			if err == io.EOF {
				// log.Info("Disconnect")
			}
			break
		}

		res := &pb.Data{}
		err = proto.Unmarshal(buf, res)
		if err != nil {
			log.Errorf("Parse data err: %s", buf)
		}
		session, ok := client.Sessions.Get(res.RequestId)
		if !ok {
			parts := strings.SplitN(client.LocalAddr, "://", 2)
			conn, err := net.Dial(parts[0], parts[1])
			if err != nil {
				log.Errorf("%s %s Dial Error", client.Name, client.LocalAddr)
				return
			}
			session = client.Sessions.Add(res.RequestId, NewConn(conn))
		}

		if string(res.Payload) == "EOF" {
			client.Sessions.Delete(res.RequestId)
		} else {
			session.Send(res.Payload)
		}
	}
}

// Close Client
func (client *Client) Close() {
	client.Sessions.Close()
}
