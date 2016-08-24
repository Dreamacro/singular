package singular

import (
	"crypto/tls"
	"net"

	pb "github.com/Dreamacro/singular/protobuf"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/satori/go.uuid"
)

// Proxy define a Proxy
type Proxy struct {
	Sessions
	Name      string
	Port      int
	Listener  net.Listener
	tls       bool
	tlsConfig *tls.Config
}

// Listen proxy listen
func (proxy *Proxy) Listen() {
	defer proxy.Close()
	go proxy.handleProxy()
	for {
		conn, err := proxy.Listener.Accept()
		if err != nil {
			break
		}
		go proxy.handleConnection(conn)
	}
}

func (proxy *Proxy) handleProxy() {
	conn := proxy.Sessions.Client
	if proxy.tls {
		conn = NewConn(tls.Server(conn.Conn, proxy.tlsConfig))
	}
	defer proxy.Close()

	for {
		buf, err := conn.Receive()
		if err != nil {
			break
		}

		res := &pb.Data{}
		err = proto.Unmarshal(buf, res)

		session, ok := proxy.Sessions.Get(res.RequestId)

		if !ok {
			continue
		}

		if string(res.Payload) == "EOF" {
			proxy.Sessions.Delete(res.RequestId)
		} else {
			session.Send(res.Payload)
		}
	}
}

// Close notification
func (proxy *Proxy) Close() {
	err := proxy.Listener.Close()
	if err != nil {
		return
	}
	proxy.Sessions.Close()
	log.Errorf("Proxy %s at %s Leave", proxy.Name, proxy.Listener.Addr())
}

func (proxy *Proxy) handleConnection(conn net.Conn) {
	log.Printf("Handle Connection: %s", conn.RemoteAddr().String())
	sessionID := uuid.NewV4().Bytes()
	c := NewConn(conn)

	proxy.Sessions.Add(sessionID, c)
}

// NewProxy create a new proxy
func NewProxy(conn Conn, name string, useTLS bool, tlsConfig *tls.Config) *Proxy {
	listen, err := net.Listen("tcp", ":0")
	CheckError("Listen TCP Error", err)

	addr, err := net.ResolveTCPAddr("tcp", listen.Addr().String())

	if useTLS {
		conn = NewConn(tls.Server(conn, tlsConfig))
	}

	return &Proxy{
		Sessions:  NewSessions(conn),
		Name:      name,
		Port:      addr.Port,
		Listener:  listen,
		tls:       useTLS,
		tlsConfig: tlsConfig,
	}
}
