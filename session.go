package singular

import (
	"io"

	pb "github.com/Dreamacro/singular/protobuf"
	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
)

// Session define session.
type Session struct {
	ID     []byte
	conn   Conn
	CloseC chan []byte
	Done   chan struct{}
}

// Sessions define sessions.
type Sessions struct {
	Sessions map[string]Session
	Client   Conn
	CloseC   chan []byte
	OutC     chan []byte
	Done     chan struct{}
}

// Add session to map
func (sessions *Sessions) Add(id []byte, sessionConn Conn) Session {
	session := NewSession(id, sessionConn, sessions.OutC, sessions.CloseC)
	sessions.Sessions[string(id)] = session
	log.Infof("New Session: %x, Session Num: %d", session.ID, len(sessions.Sessions))
	return session
}

// Delete session from map
func (sessions *Sessions) Delete(id []byte) {
	session, ok := sessions.Get(id)
	if ok {
		session.Close()
		delete(sessions.Sessions, string(id))
		log.Infof("Session Leave: %x, Session Num: %d", session.ID, len(sessions.Sessions))
	}
}

// Get session from map
func (sessions *Sessions) Get(id []byte) (session Session, ok bool) {
	session, ok = sessions.Sessions[string(id)]
	return
}

func (sessions *Sessions) handleClose() {
	for {
		select {
		case id := <-sessions.CloseC:
			sessions.Delete(id)
		case <-sessions.Done:
			return
		}
	}
}

func (sessions *Sessions) handleOut() {
	for {
		select {
		case <-sessions.Done:
			return
		case buf := <-sessions.OutC:
			sessions.Client.Send(buf)
		}
	}
}

// Close all sessions and goroutine
func (sessions *Sessions) Close() {
	for _, s := range sessions.Sessions {
		s.Close()
		log.Infof("Session Leave: %x, Session Num: %d", s.ID, len(sessions.Sessions))
	}
	sessions.Client.Close()
	close(sessions.Done)
}

// NewSession return a new session and start a goroutine
func NewSession(sessionID []byte, conn Conn, out chan []byte, close chan []byte) Session {
	session := Session{
		ID:     sessionID,
		conn:   conn,
		CloseC: close,
		Done:   make(chan struct{}),
	}
	go session.handleSession(out)
	return session
}

// NewSessions return a sessions
func NewSessions(client Conn) Sessions {
	sessions := Sessions{
		Sessions: make(map[string]Session),
		Client:   client,
		CloseC:   make(chan []byte),
		OutC:     make(chan []byte),
		Done:     make(chan struct{}),
	}
	go sessions.handleClose()
	go sessions.handleOut()
	return sessions
}

// Send data to chan
func (session *Session) Send(buf []byte) {
	session.conn.Write(buf)
}

// Close the conn
func (session *Session) Close() {
	err := session.conn.Close()
	if err != nil {
		return
	}
	close(session.Done)
	session.CloseC <- session.ID
}

func (session *Session) handleSession(out chan<- []byte) {
	defer session.Close()
	var buf = make([]byte, 1024)
	for {
		n, err := session.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				req := &pb.Data{
					RequestId: session.ID,
					Payload:   []byte("EOF"),
				}
				buffer, _ := proto.Marshal(req)
				out <- buffer
			}
			break
		}
		if n > 0 {
			// log.Infof("Rec from con: %v", buf[:n])
			req := &pb.Data{
				RequestId: session.ID,
				Payload:   buf[:n],
			}
			buffer, _ := proto.Marshal(req)

			out <- buffer
		}
	}
}
