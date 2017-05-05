package unix

import (
	"log"
	"net"
)

//UConn ...
type UConn struct {
	Query  string
	Qchan  chan string
	Rwchan chan string
}

// Server cotains the listener and Chan passing for access Query
type Server struct {
	l         *net.UnixListener
	UConnChan chan *UConn
}

//InitUnixServer ...
func InitUnixServer(conf Config) (*Server, error) {
	s := new(Server)
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: conf.Daemon.Socket, Net: "unix"})
	if err != nil {
		return nil, err
	}
	s.l = l
	s.UConnChan = make(chan *UConn)

	return s, nil
}

// write back to conn
func (s *Server) writeConn(c *net.UnixConn, u *UConn) {
	for {
		select {
		case result := <-u.Qchan:
			_, err := c.Write([]byte(result))
			if err != nil {
				log.Println("Write: ", err)
			}
		case <-u.Rwchan:
			return
		}
	}
}

// read from conn
func (s *Server) readConn(c *net.UnixConn, u *UConn) {

	buf := make([]byte, 256)
	for {
		n, err := c.Read(buf[:])
		if err != nil {
			u.Rwchan <- "close"
			buf = nil
			c.Close()
			return
		}
		u.Query = string(buf[:n])
		s.UConnChan <- u
	}
}

//Start ...
func (s *Server) Start() error {

	for {
		fd, err := s.l.AcceptUnix()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		connTable := new(UConn)
		connTable.Qchan = make(chan string)
		connTable.Rwchan = make(chan string)
		go s.writeConn(fd, connTable)
		go s.readConn(fd, connTable)
	}
}

// Stop ...
func (s *Server) Stop() error {
	return s.l.Close()
}
