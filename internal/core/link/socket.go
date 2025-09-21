package link

import (
	"net"
	"time"
)

// Constantes para comunicação via socket
const (
	defaultPort = 5738
	bufferSize  = 256
)

// SocketLink representa uma conexão via socket
type SocketLink struct {
	*Link
	server     bool
	tcpConn    net.Conn
	tcpListen  net.Listener
	serverAddr string
	serverPort int
	buffer     []byte
}

// NewSocketLink cria uma nova instância de SocketLink
func NewSocketLink(server bool, addr string, port int) *SocketLink {
	if port == 0 {
		port = defaultPort
	}
	if addr == "" {
		addr = "localhost"
	}

	return &SocketLink{
		Link:       NewLink(),
		server:     server,
		serverAddr: addr,
		serverPort: port,
		buffer:     make([]byte, bufferSize),
	}
}

// Connect estabelece a conexão via socket
func (s *SocketLink) Connect() ConnectionState {
	if s.server {
		return s.startServer()
	}
	return s.connectToServer()
}

// startServer inicia o servidor TCP
func (s *SocketLink) startServer() ConnectionState {
	var err error
	addr := net.JoinHostPort(s.serverAddr, string(rune(s.serverPort)))
	s.tcpListen, err = net.Listen("tcp", addr)
	if err != nil {
		return LINK_ERROR
	}

	// Aceita conexões em uma goroutine
	go func() {
		for {
			conn, err := s.tcpListen.Accept()
			if err != nil {
				continue
			}
			s.tcpConn = conn
			s.connectedSlaves++
		}
	}()

	return LINK_OK
}

// connectToServer conecta ao servidor TCP
func (s *SocketLink) connectToServer() ConnectionState {
	var err error
	addr := net.JoinHostPort(s.serverAddr, string(rune(s.serverPort)))
	s.tcpConn, err = net.Dial("tcp", addr)
	if err != nil {
		return LINK_ERROR
	}

	return LINK_OK
}

// Send envia dados para a conexão
func (s *SocketLink) Send(data []byte) error {
	if s.tcpConn == nil {
		return nil
	}

	_, err := s.tcpConn.Write(data)
	return err
}

// Receive recebe dados da conexão
func (s *SocketLink) Receive(timeout time.Duration) ([]byte, error) {
	if s.tcpConn == nil {
		return nil, nil
	}

	// Define timeout para a leitura
	s.tcpConn.SetReadDeadline(time.Now().Add(timeout))

	n, err := s.tcpConn.Read(s.buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, nil
		}
		return nil, err
	}

	return s.buffer[:n], nil
}

// Close fecha a conexão
func (s *SocketLink) Close() {
	if s.tcpConn != nil {
		s.tcpConn.Close()
		s.tcpConn = nil
	}
	if s.tcpListen != nil {
		s.tcpListen.Close()
		s.tcpListen = nil
	}
}

// UpdateSocket atualiza o estado da conexão via socket
func (s *SocketLink) UpdateSocket(ticks int64) {
	s.lastUpdate += ticks

	if !s.enabled || !s.transferring {
		return
	}

	// TODO: Implementar lógica de atualização específica para socket
	// - Verificar dados recebidos
	// - Atualizar estado da transferência
	// - Gerenciar timeouts
}
