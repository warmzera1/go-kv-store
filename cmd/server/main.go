package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/warmzera1/kv-store/internal/server"
	"github.com/warmzera1/kv-store/internal/store/core"
	"github.com/warmzera1/kv-store/internal/store/pkg/ttl"
)

func main() {
	fmt.Println("Starting KV-Store server...")

	// 1. Создаем хранилище
	s := core.New()

	// 2. Загружаем сохраненные данные
	err := s.LoadSnapshot("test.gob")
	if err != nil {
		fmt.Printf("Warning: could not load snapshot: %v\n", err)
	} else {
		fmt.Printf("Loaded shapshot from test.gob\n")
	}

	// 3. Запускаем TTL очистку
	ttlStore := ttl.New(s)
	ttlStore.StartTTLCleaner(1 * time.Second)
	fmt.Println("TTL cleaner started")

	// 4. Создаем TCP сервер
	svr := server.New(":6379", s)

	// 5. Настраиваем graceful shutdown
	// 5.1 Создаем канал для сигналов
	// sigChan - это канал для передачи сигналов
	// Вместимость 1, значит можем сохранить 1 сигнал в очереди
	sigChan := make(chan os.Signal, 1)

	// 6. Подписываемся на сигналы
	// SIGINT - Cntrl + C
	// SIGTERM - команда kill
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 6.1 Запускаем горутину слушателя
	go func() {
		// Ждем сигнал (блокируемся, пока канал не получит значение)
		<-sigChan // Читаем из канала, ждем
		fmt.Println("\nShutdown down server...")

		// Сохраняем при выходе
		err := s.SaveSnapshot("test.gob")
		if err != nil {
			fmt.Printf("Error saving snapshot: %v\n", err)
		} else {
			fmt.Printf("Snapshot saved")
		}

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
