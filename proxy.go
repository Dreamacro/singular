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
	Port      int
	Listener  net.Listener
	Alive     bool
	tls       bool
	tlsConfig *tls.Config
}

// Listen proxy listen
func (proxy *Proxy) Listen() {
	defer proxy.Listener.Close()
	for {
		conn, err := proxy.Listener.Accept()
		if err != nil {
			log.Errorf("Proxy Leave")
			break
		}
		if proxy.Alive {
			go proxy.handleConnection(conn)
		} else {
			go proxy.handleProxy(conn)
		}
	}
}

func (proxy *Proxy) handleProxy(conn net.Conn) {
	if proxy.tls {
		conn = tls.Server(conn, proxy.tlsConfig)
	}
	proxy.Alive = true
	proxy.Sessions = NewSessions(NewConn(conn))
	defer proxy.Close()

	for {
		buf, err := proxy.Client.Receive()
		if err != nil {
			// log.Info("Disconnect")

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
}

func (proxy *Proxy) handleConnection(conn net.Conn) {
	log.Printf("Handle Connection: %s", conn.RemoteAddr().String())
	sessionID := uuid.NewV4().Bytes()
	c := NewConn(conn)

	proxy.Sessions.Add(sessionID, c)
}

// NewProxy create a new proxy
func NewProxy(useTLS bool, tlsConfig *tls.Config) *Proxy {
	listen, err := net.Listen("tcp", ":0")
	CheckError("Listen TCP Error", err)

	addr, err := net.ResolveTCPAddr("tcp", listen.Addr().String())

	return &Proxy{
		Port:      addr.Port,
		Listener:  listen,
		Alive:     false,
		tls:       useTLS,
		tlsConfig: tlsConfig,
	}
}
