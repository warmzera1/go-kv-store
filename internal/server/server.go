package server

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

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
	}

}

// executeCommand - разбирает и выполняет команду
func (s *Server) executeCommand(parts []string) string {
	// Разбираем команду на слова
	if len(parts) == 0 {
		return "ERROR: empty command"
	}

	// Приводим команду к верхнему регистру
	command := strings.ToUpper(parts[0])

	switch command {

	// 1. PING - проверка соединения
	case "PING":
		return protocol.EncodeSimpleString("PONG")

	// 2. QUIT - закрытие соединения
	case "QUIT":
		return protocol.EncodeSimpleString("OK")

	// 2. SET - установить значение
	case "SET":
		// Проверяем количество аргументов
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: SET requires key and value")
		}
		key := parts[1]
		// Объединяем все остальные слова в значение (если значение содержит пробелы)
		value := strings.Join(parts[2:], " ")
		s.stringCmd.Set(key, value)
		return protocol.EncodeSimpleString("OK")

	// 3. GET - получить значение
	case "GET":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: GET requires key")
		}
		key := parts[1]
		value, exists := s.stringCmd.Get(key)
		if !exists {
			return protocol.EncodeNullBulkString()
		}
		return protocol.EncodeBulkString(value)

	// 4. DEL - удалить ключ
	case "DEL":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: DEL requires key")
		}
		key := parts[1]
		s.stringCmd.Delete(key)
		return protocol.EncodeSimpleString("OK")

	// 5. Expire - установить время жизни ключа
	case "EXPIRE":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: EXPIRE requires key and seconds")
		}
		key := parts[1]

		seconds, err := strconv.Atoi(parts[2])
		if err != nil {
			return protocol.EncodeError("ERR value is not an integer")
		}

		ok, err := s.ttlCmd.Expire(key, seconds)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}
		if ok {
			return protocol.EncodeInteger(1)
		}
		return protocol.EncodeInteger(0)

	// 6. TTL - время жизни ключа
	case "TTL":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: TTL requires key")
		}
		key := parts[1]

		ttl, err := s.ttlCmd.TTL(key)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		// -2 нет ключа, -1 нет TTL, >0 секунды
		return protocol.EncodeInteger(ttl)

	// 7. PERSIST - удаляет TTL у ключа
	case "PERSIST":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: PERSIST requires key")
		}
		key := parts[1]

		ok, err := s.ttlCmd.Persist(key)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}
		if ok {
			// TTL - удален
			return protocol.EncodeInteger(1)
		}

		// TTL - не был установлен или нет ключа
		return protocol.EncodeInteger(0)

	// 8. LPUSH - добавляет элемент в начало списка
	case "LPUSH":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: LPUSH requires key and value(s)")
		}
		key := parts[1]
		values := parts[2:]

		// Преобразовываем []string в []interface{}
		interfaceValues := make([]interface{}, len(values))
		for i, v := range values {
			interfaceValues[i] = v
		}

		// Выполняем LPush
		err := s.listCmd.LPush(key, interfaceValues...)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		// Получаем и возвращаем новую длину списка
		lenght, _ := s.listCmd.LLen(key)
		return protocol.EncodeInteger(lenght)

	// 9. RPush - добавляет элемент в конец списка
	case "RPUSH":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: RPUSH requires key and value(s)")
		}
		key := parts[1]
		values := parts[2:]

		interfaceValues := make([]interface{}, len(values))
		for i, v := range values {
			interfaceValues[i] = v
		}

		// Выполняем RPush
		err := s.listCmd.RPush(key, interfaceValues...)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		// Получаем и возвращаем новую длину списка
		length, _ := s.listCmd.LLen(key)
		return protocol.EncodeInteger(length)

	// 10. LPop - удаляет первый элемент
	case "LPOP":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: LPOP requires key")
		}
		key := parts[1]
		value, err := s.listCmd.LPop(key)
		if err != nil {
			return protocol.EncodeNullBulkString()
		}
		return protocol.EncodeBulkString(fmt.Sprintf("%v", value))

	// 11. RPop - удаляет последний элемент
	case "RPOP":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: RPOP requires key")
		}
		key := parts[1]
		value, err := s.listCmd.RPop(key)
		if err != nil {
			return protocol.EncodeNullBulkString()
		}
		return protocol.EncodeBulkString(fmt.Sprintf("%v", value))

	// LLen - возвращает длину списка
	case "LLEN":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: LLEN requires key")
		}
		key := parts[1]
		length, err := s.listCmd.LLen(key)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}
		return protocol.EncodeInteger(length)

	// LIndex - возвращает элемент по индексу
	case "LINDEX":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: LINDEX requires key and index")
		}
		key := parts[1]
		index, err := strconv.Atoi(parts[2])
		if err != nil {
			return protocol.EncodeError("ERROR: invalid index")
		}
		value, err := s.listCmd.LIndex(key, index)
		if err != nil {
			return protocol.EncodeNullBulkString()
		}

		return protocol.EncodeBulkString(fmt.Sprintf("%v", value))

	// LRange - возвращает диапазон элементов
	case "LRANGE":
		if len(parts) < 4 {
			return protocol.EncodeError("ERROR: LRANGE requires key, start and stop")
		}
		key := parts[1]
		start, err := strconv.Atoi(parts[2])
		if err != nil {
			return protocol.EncodeError("ERROR: invalid start")
		}
		stop, err := strconv.Atoi(parts[3])
		if err != nil {
			return protocol.EncodeError("ERROR: invalid stop")
		}

		items, err := s.listCmd.LRange(key, start, stop)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}
		if len(items) == 0 {
			return protocol.EncodeArray([]string{})
		}

		result := make([]string, len(items))
		for i, v := range items {
			result[i] = fmt.Sprintf("%v", v)
		}

		return protocol.EncodeArray(result)

	// SADD - добавляет элемент в множество
	case "SADD":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: SADD requires key and value(s)")
		}
		key := parts[1]
		members := parts[2:]

		membersInterface := make([]interface{}, len(members))
		for k, v := range members {
			membersInterface[k] = v
		}

		added, err := s.setCmd.SAdd(key, membersInterface...)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		return protocol.EncodeInteger(added)

		// SREM - удаляет один или несколько элементов
	case "SREM":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: SREM requires key and value(s)")
		}
		key := parts[1]
		members := parts[2:]

		membersInterface := make([]interface{}, len(members))
		for k, v := range members {
			membersInterface[k] = v
		}

		removed, err := s.setCmd.SRem(key, membersInterface...)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		return protocol.EncodeInteger(removed)

	// SISMEMBER - есть ли элемент в множестве
	case "SISMEMBER":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: SISMEMBER requires key and value")
		}
		key := parts[1]
		value := parts[2]

		exists, err := s.setCmd.SIsMember(key, value)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}
		if exists {
			return protocol.EncodeInteger(1)
		}

		return protocol.EncodeInteger(0)

	// SMEMBERS - возвращает все элементы в множестве
	case "SMEMBERS":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: SMEMBERS requires key")
		}
		key := parts[1]

		members, err := s.setCmd.SMembers(key)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}
		if len(members) == 0 {
			return protocol.EncodeArray([]string{})
		}

		result := make([]string, len(members))
		for i, v := range members {
			result[i] = fmt.Sprintf("%v", v)
		}

		return protocol.EncodeArray(result)

	// SCARD - возвращает кол-во элементов в множестве
	case "SCARD":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: SCARD requires key")
		}
		key := parts[1]

		count, err := s.setCmd.SCard(key)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		return protocol.EncodeInteger(count)

	// HSET - устанавливает поле в хеше
	case "HSET":
		if len(parts) < 4 {
			return protocol.EncodeError("ERROR: HSET requires key, field and value")
		}
		key := parts[1]
		field := parts[2]
		value := strings.Join(parts[3:], " ")

		err := s.hashCmd.HSet(key, field, value)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		return protocol.EncodeSimpleString("OK")

	// HGET - получить поле поключу
	case "HGET":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: HGET requires key and field")
		}
		key := parts[1]
		field := parts[2]

		value, err := s.hashCmd.HGet(key, field)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		if value == nil {
			return protocol.EncodeNullBulkString()
		}

		return protocol.EncodeBulkString(fmt.Sprintf("%v", value))

	// HGETALL - получить все поля по ключу
	case "HGETALL":
		if len(parts) < 2 {
			return protocol.EncodeError("ERROR: HGETALL requires key")
		}

		key := parts[1]
		fields, err := s.hashCmd.HGetAll(key)

		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		if len(fields) == 0 {
			return protocol.EncodeArray([]string{})
		}

		// Форматируем: каждое поле и значение с новой строки
		result := make([]string, 0, len(fields)*2)
		for k, v := range fields {
			result = append(result, k, fmt.Sprintf("%v", v))
		}

		return protocol.EncodeArray(result)

	// HDEL - удаляет одно или несколько полей из хеша
	case "HDEL":
		if len(parts) < 3 {
			return protocol.EncodeError("ERROR: HDEL requires key and field(s)")
		}
		key := parts[1]
		fields := parts[2:]

		err := s.hashCmd.HDel(key, fields...)
		if err != nil {
			return protocol.EncodeError(err.Error())
		}

		return protocol.EncodeSimpleString("OK")

	// SAVE - сохранить снапшот
	case "SAVE":
		filename := "test.gob"
		if len(parts) > 1 {
			filename = parts[1]
		}

		if err := s.adminCmd.SaveSnapshot(filename); err != nil {
			return protocol.EncodeError(err.Error())
		}

		return protocol.EncodeSimpleString("OK")

	// LOAD - загрузить файл
	case "LOAD":
		filename := "test.gob"
		if len(parts) > 1 {
			filename = parts[1]
		}

		if err := s.adminCmd.LoadSnapshot(filename); err != nil {
			return protocol.EncodeError(err.Error())
		}

		return protocol.EncodeSimpleString("OK")

	// Неизвестная команда
	default:
		return protocol.EncodeError(fmt.Sprintf("ERROR: unknown command '%s'", command))
	}
}
