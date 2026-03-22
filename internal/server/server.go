package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/warmzera1/kv-store/internal/store"
)

// Server - представляет ТСР сервер
type Server struct {
	addr     string       // адрес для прослушивания (например:6379)
	store    *store.Store // ссылка на хранилище
	listener net.Listener
	running  bool
}

// New - создает новый сервер
func New(addr string, store *store.Store) *Server {
	return &Server{
		addr:  addr,
		store: store,
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
	// 1. Закрываем соединение при выходе
	defer conn.Close()

	// conn.RemoteAddr - адрес клиента (например, 127.0.0.1:5423)
	fmt.Printf("New connection from %s\n", conn.RemoteAddr())

	// Создаем сканер для чтения команд построчно
	scanner := bufio.NewScanner(conn)

	// Читаем команды, пока соединение окрыто
	for scanner.Scan() {

		// Получаем команду (убираем лишние пробелы)
		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			continue
		}

		fmt.Printf("Received from %s: %s\n", conn.RemoteAddr(), cmd)

		// Выполняем команду и получаем ответ
		response := s.executeCommand(cmd)

		// Отправляем ответ клиенту
		_, err := conn.Write([]byte(response + "\n"))
		if err != nil {
			fmt.Printf("Error writing to %s: %v\n", conn.RemoteAddr(), err)
			break
		}
	}

	// Проверяем ошибки сканера
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from %s: %v\n", conn.RemoteAddr(), err)
	}

	fmt.Printf("Connection from %s closed\n", conn.RemoteAddr())
}

// executeCommand - разбирает и выполняет команду
func (s *Server) executeCommand(cmd string) string {
	// Разбираем команду на слова
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "ERROR: empty command"
	}

	// Приводим команду к верхнему регистру
	command := strings.ToUpper(parts[0])

	switch command {

	// 1. PING - проверка соединение
	case "PING":
		return "PONG"

	// 2. SET - установить значение
	case "SET":
		// Проверяем количество аргументов
		if len(parts) < 3 {
			return "ERROR: SET requires key and value"
		}
		key := parts[1]
		// Объединяем все остальные слова в значение (если значение содержит пробелы)
		value := strings.Join(parts[2:], " ")
		s.store.Set(key, value)
		return "OK"

	// 3. GET - получить значение
	case "GET":
		if len(parts) < 2 {
			return "ERROR: GET requires key"
		}
		key := parts[1]
		value, exists := s.store.Get(key)
		if !exists {
			return "(nil)"
		}
		return value

	// 4. DEL - удалить ключ
	case "DEL":
		if len(parts) < 2 {
			return "ERROR: DEL requires key"
		}
		key := parts[1]
		s.store.Delete(key)
		return "OK"

	case "QUIT":
		return "OK"

	// Неизвестная команда
	default:
		return fmt.Sprintf("ERROR: unknown command '%s'", command)
	}
}
