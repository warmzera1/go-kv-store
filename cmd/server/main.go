package main

// import (
// 	"fmt"
// 	"time"

// 	"github.com/warmzera1/kv-store/internal/store"
// )

// func main() {
// 	s := store.New()

// 	// Запускаем фоновую горутину для активного удаления (если реализовали)
// 	// s.StartTTLCleaner(1 * time.Second)

// 	fmt.Println("=== Тест TTL в Get ===")
// 	testTTLinGet(s)

// }

// func testTTLinGet(s *store.Store) {
// 	// 1. Устанавливаем ключ с TTL 3 секунды
// 	fmt.Println("1. Set('temp', 'value') и Expire 3 секунды")
// 	s.Set("temp", "value")
// 	s.Expire("temp", 3)

// 	// 2. Сразу проверяем - ключ должен быть
// 	val, ok := s.Get("temp")
// 	fmt.Printf("   Сразу после установки: val='%v', ok=%v\n", val, ok)

// 	// 3. Проверяем TTL
// 	ttl, _ := s.TTL("temp")
// 	fmt.Printf("   TTL: %d секунд\n", ttl)

// 	// 4. Ждем 2 секунды (ключ еще жив)
// 	fmt.Println("   Ждем 2 секунды...")
// 	time.Sleep(2 * time.Second)

// 	val, ok = s.Get("temp")
// 	fmt.Printf("   После 2 секунд: val='%v', ok=%v\n", val, ok)
// 	ttl, _ = s.TTL("temp")
// 	fmt.Printf("   TTL: %d секунд\n", ttl)

// 	// 5. Ждем еще 2 секунды (ключ должен истечь)
// 	fmt.Println("   Ждем еще 2 секунды...")
// 	time.Sleep(2 * time.Second)

// 	val, ok = s.Get("temp")
// 	fmt.Printf("   После 4 секунд: val='%v', ok=%v\n", val, ok)

// 	// Проверяем TTL (должен вернуть -2)
// 	ttl, _ = s.TTL("temp")
// 	fmt.Printf("   TTL: %d (ожидается -2 - ключ не существует)\n", ttl)

// }
