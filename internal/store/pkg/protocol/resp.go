package protocol

import (
	"bufio"
	"errors"
	"strconv"
)

// Decoder - декодирует RESP команды из reader
type Decoder struct {
	reader *bufio.Reader
}

// NewDecoder - создает новый декодер
func NewDecoder(r *bufio.Reader) *Decoder {
	return &Decoder{reader: r}
}

// Decode - читает и декодирует RESP в команду
func (d *Decoder) Decode() ([]string, error) {
	// 1. Читаем первый байт - он определяет тип
	b, err := d.reader.ReadByte()
	if err != nil {
		return nil, nil
	}

	// 2. Если это массив (начинается с '*')
	// Например: *3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n
	switch b {

	case '*':
		return d.decodeArray()
	default:
		return nil, errors.New("invalid RESP: expected array")
	}
}

// decodeArray - читает массив RESP
// Формат: *3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n
func (d *Decoder) decodeArray() ([]string, error) {
	// 1. Читаем размер массива
	// Читаем строку до '\n'
	line, err := d.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// 2. Убираем "\r\n" и преобразуем в число
	// size = 2
	size, err := strconv.Atoi(line[:len(line)-2])
	if err != nil {
		return nil, err
	}

	// 3. Создаем срез для результатов
	result := make([]string, size)

	// 3. Читаем каждый элемент массива
	// i = 0: читаем "$3\r\nSET\r\n" -> "SET"
	// i = 1: читаем "$4\r\nname\r\n" -> "name"
	// i = 2: читаем "$5\r\nAlice\r\n" -> "Alice"
	for i := 0; i < size; i++ {
		// Каждый элемент это bulk string (начинается с $)
		item, err := d.decodeBulkString()
		if err != nil {
			return nil, err
		}
		result[i] = item
	}

	// ["SET", "name", "Alice"]
	return result, nil
}

// decodeBulkString - читает bulk string
// Форматы:
// - обычная строка: $5\r\nhello\r\n
// - Null строка: $-1\r\n
func (d *Decoder) decodeBulkString() (string, error) {
	// 1. Читаем строку с длиной (начинается с "$")
	// $5\r\n
	line, err := d.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// 2. Убираем "$" и "\r\n"
	// line = "$"
	lengthStr := line[1 : len(line)-2]

	// 3. Преобразуем длину в число
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", err
	}

	// 4. Проверяем на null строку ($-1\r\n)
	if length == -1 {
		return "", nil
	}

	// 5. Читаем саму строку + "\r\n"
	// Нужно прочитать length + 2 байта "\r\n"
	buf := make([]byte, length+2)
	_, err = d.reader.Read(buf)
	if err != nil {
		return "", err
	}

	// 6. Возвращаем строку без "\r\n"
	return string(buf[:length]), nil
}

// EncodeSimpleString - кодирует простую строку (+OK\r\n)
func EncodeSimpleString(s string) string {
	return "+" + s + "\r\n"
}

// EncodeError - кодирует ошибку (-ERROR\r\n)
func EncodeError(err string) string {
	return "-" + err + "\r\n"
}

// EncodeInteger - кодирует число (:100\r\n)
func EncodeInteger(n int) string {
	return ":" + strconv.Itoa(n) + "\r\n"
}

// EncodeBulkString - кодирует bulk string ($5\r\nhello\r\n)
func EncodeBulkString(s string) string {
	if s == "" {
		return "$-1\r\n"
	}
	return "$" + strconv.Itoa(len(s)) + "\r\n" + s + "\r\n"
}

// EncodeNulBulkString - кодирует null ($-1\r\n)
func EncodeNullBulkString() string {
	return "$-1\r\n"
}

// EncodeArray - кодирует массив (*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n)
func EncodeArray(items []string) string {
	result := "*" + strconv.Itoa(len(items)) + "\r\n"
	for _, item := range items {
		result += EncodeBulkString(item)
	}

	return result
}
