package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/warmzera1/kv-store/internal/server"
	"github.com/warmzera1/kv-store/internal/store/core"
)

func main() {
	fmt.Println("Starting KV-Store server...")

	// 1. Создаем хранилище
	s := core.New()

	// // 2. Запускаем TTL
	// s.StartTTLCleaner(1 * time.Second)

	// 3. Создаем TCP сервер
	svr := server.New(":6379", s)

	// 4. Настраиваем graceful shutdown
	// 4.1 Создаем канал для сигналов
	// sigChan - это канал для передачи сигналов
	// Вместимость 1, значит можем сохранить 1 сигнал в очереди
	sigChan := make(chan os.Signal, 1)

	// 5. Подписываемся на сигналы
	// 	SIGINT - Cntrl + C
	// SIGTERM - команда kill
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 5.1 Запускаем горутину слушателя
	go func() {
		// Ждем сигнал (блокируемся, пока канал не получит значение)
		<-sigChan // Читаем из канала, ждем
		fmt.Println("\nShutdown down server...")
		if err := svr.Stop(); err != nil {
			fmt.Printf("Error stoping server: %v\n", err)
		}

		// Завершаем программу
		os.Exit(0)
	}()

	// 6. Запускаем сервер (блокируем выполнение)
	if err := svr.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
