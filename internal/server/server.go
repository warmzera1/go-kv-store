package server

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
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

	// 1. PING - проверка соединения
	case "PING":
		return "PONG"

	// 2. QUIT - закрытие соединения
	case "QUIT":
		return "OK"

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

	// 5. Expire - установить время жизни ключа
	case "EXPIRE":
		if len(parts) < 3 {
			return "ERROR: EXPIRE requires key and seconds"
		}
		key := parts[1]
		seconds := 0
		fmt.Sscanf(parts[2], "%d", &seconds)
		ok, err := s.store.Expire(key, seconds)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}
		if ok {
			return "1"
		}
		return "0"

	// 6. TTL - время жизни ключа
	case "TTL":
		if len(parts) < 2 {
			return "ERROR: TTL requires key"
		}
		key := parts[1]

		ttl, err := s.store.TTL(key)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		// -2 нет ключа, -1 нет TTL, >0 секунды
		return fmt.Sprintf("%d", ttl)

	// 7. PERSIST - удаляет TTL у ключа
	case "PERSIST":
		if len(parts) < 2 {
			return "ERROR: PERSIST requires key"
		}
		key := parts[1]

		ok, err := s.store.Persist(key)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}
		if ok {
			// TTL - удален
			return "1"
		}

		// TTL - не был установлен или нет ключа
		return "0"

	// 8. LPUSH - добавляет элемент в начало списка
	case "LPUSH":
		if len(parts) < 3 {
			return "ERROR: LPUSH requires key and value(s)"
		}
		key := parts[1]
		values := parts[2:]

		// Преобразовываем []string в []interface{}
		interfaceValues := make([]interface{}, len(values))
		for i, v := range values {
			interfaceValues[i] = v
		}

		// Выполняем LPush
		err := s.store.LPush(key, interfaceValues...)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		// Получаем и возвращаем новую длину списка
		lenght, _ := s.store.LLen(key)
		return fmt.Sprintf("%d", lenght)

	// 9. RPush - добавляет элемент в конец списка
	case "RPUSH":
		if len(parts) < 3 {
			return "ERROR: RPUSH requires key and value(s)"
		}
		key := parts[1]
		values := parts[2:]

		interfaceValues := make([]interface{}, len(values))
		for i, v := range values {
			interfaceValues[i] = v
		}

		// Выполняем RPush
		err := s.store.RPush(key, interfaceValues...)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		// Получаем и возвращаем новую длину списка
		lenght, _ := s.store.LLen(key)
		return fmt.Sprintf("%d", lenght)

	// 10. LPop - удаляет первый элемент
	case "LPOP":
		if len(parts) < 2 {
			return "ERROR: LPOP requires key"
		}
		key := parts[1]
		value, err := s.store.LPop(key)
		if err != nil {
			return "(nil)"
		}
		return fmt.Sprintf("%v", value)

	// 11. RPop - удаляет последний элемент
	case "RPOP":
		if len(parts) < 2 {
			return "ERROR: RPOP requires key"
		}
		key := parts[1]
		value, err := s.store.RPop(key)
		if err != nil {
			return "(nil)"
		}
		return fmt.Sprintf("%v", value)

	// LLen - возвращает длину списка
	case "LLEN":
		if len(parts) < 2 {
			return "ERROR: LLEN requires key"
		}
		key := parts[1]
		length, err := s.store.LLen(key)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}
		return fmt.Sprintf("%d", length)

	// LIndex - возвращает элемент по индексу
	case "LINDEX":
		if len(parts) < 3 {
			return "ERROR: LINDEX requires key and index"
		}
		key := parts[1]
		index, err := strconv.Atoi(parts[2])
		if err != nil {
			return "ERROR: invalid index"
		}
		value, err := s.store.LIndex(key, index)
		if err != nil {
			return "(nil)"
		}

		return fmt.Sprintf("%v", value)

	// LRange - возвращает диапазон элементов
	case "LRANGE":
		if len(parts) < 4 {
			return "ERROR: LRANGE requires key, start and stop"
		}
		key := parts[1]
		start, err := strconv.Atoi(parts[2])
		if err != nil {
			return "ERROR: invalid start"
		}
		stop, err := strconv.Atoi(parts[3])
		if err != nil {
			return "ERROR: invalid stop"
		}

		items, err := s.store.LRange(key, start, stop)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}
		if len(items) == 0 {
			return "empty array"
		}

		result := make([]string, len(items))
		for i, v := range items {
			result[i] = fmt.Sprintf("%v", v)
		}

		return strings.Join(result, "\n")

	// SADD - добавляет элемент в множество
	case "SADD":
		if len(parts) < 3 {
			return "ERROR: SADD requires key and value(s)"
		}
		key := parts[1]
		members := parts[2:]

		membersInterface := make([]interface{}, len(members))
		for k, v := range members {
			membersInterface[k] = v
		}

		added, err := s.store.SAdd(key, membersInterface...)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		return fmt.Sprintf("%d", added)

		// SREM - удаляет один или несколько элементов
	case "SREM":
		if len(parts) < 3 {
			return "ERROR: SREM requires key and value(s)"
		}
		key := parts[1]
		members := parts[2:]

		membersInterface := make([]interface{}, len(members))
		for k, v := range members {
			membersInterface[k] = v
		}

		removed, err := s.store.SRem(key, membersInterface...)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		return fmt.Sprintf("%d", removed)

	// SISMEMBER - есть ли элемент в множестве
	case "SISMEMBER":
		if len(parts) < 3 {
			return "ERROR: SISMEMBER requires key and value"
		}
		key := parts[1]
		value := parts[2]

		exists, err := s.store.SIsMember(key, value)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}
		if exists {
			return "1"
		}

		return "0"

	// SMEMBERS - возвращает все элементы в множестве
	case "SMEMBERS":
		if len(parts) < 2 {
			return "ERROR: SMEMBERS requires key"
		}
		key := parts[1]

		members, err := s.store.SMembers(key)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}
		if len(members) == 0 {
			return "(empty set)"
		}

		result := make([]string, len(members))
		for i, v := range members {
			result[i] = fmt.Sprintf("%v", v)
		}

		return strings.Join(result, "\n")

	// SCARD - возвращает кол-во элементов в множестве
	case "SCARD":
		if len(parts) < 2 {
			return "ERROR: SCARD requires key"
		}
		key := parts[1]

		count, err := s.store.SCard(key)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		return fmt.Sprintf("%d", count)

	// HSET - устанавливает поле в хеше
	case "HSET":
		if len(parts) < 4 {
			return "ERROR: HSET requires key, field and value"
		}
		key := parts[1]
		field := parts[2]
		value := strings.Join(parts[3:], " ")

		err := s.store.HSet(key, field, value)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		return "OK"

	// HGET - получить поле поключу
	case "HGET":
		if len(parts) < 3 {
			return "ERROR: HGET requires key and field"
		}
		key := parts[1]
		field := parts[2]

		value, err := s.store.HGet(key, field)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		if value == nil {
			return "(nil)"
		}

		return fmt.Sprintf("%v", value)

	// HGETALL - получить все поля по ключу
	case "HGETALL":
		if len(parts) < 2 {
			return "ERROR: HGETALL requires key"
		}

		key := parts[1]
		fields, err := s.store.HGetAll(key)

		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		if len(fields) == 0 {
			return "(empty hash)"
		}

		// Форматируем: каждое поле и значение с новой строки
		result := make([]string, 0, len(fields)*2)
		for k, v := range fields {
			result = append(result, k, fmt.Sprintf("%v", v))
		}

		return strings.Join(result, "\n")

	// HDEL - удаляет одно или несколько полей из хеша
	case "HDEL":
		if len(parts) < 3 {
			return "ERROR: HDEL requires key and field(s)"
		}
		key := parts[1]
		fields := parts[2:]

		err := s.store.HDel(key, fields...)
		if err != nil {
			return fmt.Sprintf("ERROR: %v", err)
		}

		return "OK"

	// Неизвестная команда
	default:
		return fmt.Sprintf("ERROR: unknown command '%s'", command)
	}

}
