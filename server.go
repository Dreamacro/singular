package singular

import (
	"crypto/tls"
	"fmt"
	"net"

	pb "github.com/Dreamacro/singular/protobuf"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

// Server define a server type
type Server struct {
	tls       bool
	tlsConfig *tls.Config
}

// NewServer generator a new server
func NewServer(tls bool, tlsConfig *tls.Config) *Server {
	return &Server{
		tls:       tls,
		tlsConfig: tlsConfig,
	}
}

// Serve serve a adderss
func (server *Server) Serve(port int) {
	addr := fmt.Sprintf(":%d", port)
	var listen net.Listener
	var err error
	if server.tls {
		listen, err = tls.Listen("tcp", addr, server.tlsConfig)
	} else {
		listen, err = net.Listen("tcp", addr)
	}

	PassOrFatal("Listen Error", err)
	defer listen.Close()
	log.Printf("Server started on %s", listen.Addr())

	for {
		conn, err := listen.Accept()
		CheckError("Listen Accept Error", err)

		go server.handleClient(conn)
	}
}

func (server *Server) handleClient(conn net.Conn) {
	connection := NewConn(conn)
	defer connection.Close()
	log.Infof("New Client Connection: %s", connection.RemoteAddr().String())
	buf, err := connection.Receive()
	CheckError("Client Message Error", err)

	req := &pb.Request{}
	err = proto.Unmarshal(buf, req)
	CheckError("Unmarshal Message Error", err)
	if req.Meta == pb.Request_NewProxy {
		proxy := NewProxy(server.tls, server.tlsConfig)

		req := &pb.Request{
			Meta:    pb.Request_Assign,
			Payload: fmt.Sprintf("%d", proxy.Port),
		}
		log.Infof("Assign Proxy: %s", proxy.Listener.Addr())
		buf, err := proto.Marshal(req)
		if err != nil {
			log.Errorf("buf error: %v", err)
		}
		connection.Send(buf)

		proxy.Listen()
	}
}
