package server

import (
	"bufio"
	"fmt"
	"net"

	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/admin"
	"github.com/warmzera1/kv-store/internal/store/pkg/hash"
	"github.com/warmzera1/kv-store/internal/store/pkg/list"
	"github.com/warmzera1/kv-store/internal/store/pkg/protocol"
	"github.com/warmzera1/kv-store/internal/store/pkg/set"
	str "github.com/warmzera1/kv-store/internal/store/pkg/string"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

// Server - представляет ТСР сервер
type Server struct {
	addr      string // адрес для прослушивания (например:6379)
	stringCmd str.InterfaceString
	listCmd   list.ListInterface
	setCmd    set.SetInterface
	hashCmd   hash.HashInterface
	ttlCmd    ttl.TTLInterface
	adminCmd  admin.AdminInterface
	listener  net.Listener
	running   bool
	quit      bool
}

// New - создает новый сервер
func New(addr string, s *core.Store) *Server {
	ttlStore := ttl.New(s)

	return &Server{
		addr:      addr,
		stringCmd: str.New(s, ttlStore), // создаем обертку
		listCmd:   list.New(s, ttlStore),
		setCmd:    set.New(s, ttlStore),
		hashCmd:   hash.New(s, ttlStore),
		adminCmd:  admin.New(s),
		ttlCmd:    ttlStore,
	}
}

// Start - запускает сервер
func (s *Server) Start() error {
	// 1. Создаем listener
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	s.listener = listener
	s.running = true

	fmt.Printf("KV-Store server listening on %s\n", s.addr)

	// 2. Основной цикл принятия сервера
	for s.running {
		// Принимаем новое соединение (блокирующая операция)
		conn, err := listener.Accept()
		if err != nil {
			// Если сервер закрыт, выходим
			if !s.running {
				return nil
			}
			fmt.Printf("Error accepting connection %v\n", err)
			continue
		}

		// 3. Каждое соединение обрабатываем в отдельной горутине
		// Каждый клиент обрабатываем параллельно
		go s.handleConnection(conn)
	}

	return nil
}

// Stop - останавливает сервер
func (s *Server) Stop() error {
	s.running = false
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	decoder := protocol.NewDecoder(reader)

	for {
		parts, err := decoder.Decode()
		if err != nil {
			conn.Write([]byte(protocol.EncodeError("ERR invalid command")))
			break
		}

		if len(parts) == 0 {
			continue
		}

		response := s.executeCommand(parts)
		conn.Write([]byte(response))
		if s.quit {
			return
		}
	}

}
