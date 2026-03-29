package core

import (
	"encoding/gob"
	"os"
	"time"
)

func init() {
	gob.Register(&ListValue{})
	gob.Register(&SetValue{})
	gob.Register(&HashValue{})
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register(time.Time{})
}

type SnapshotData struct {
	Data   map[string]interface{} `json:"data"`
	Types  map[string]DataType    `json:"types"`
	Expiry map[string]time.Time   `json:"expiry"`
	Stats  Stats                  `json:"stats"`
}

// SaveSnapshot - сохраняет текущее состояние в файл
func (s *Store) SaveSnapshot(filename string) error {
	// 1. Блокируем для чтения
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	// 2. Собираем данные для сохранения
	snap := SnapshotData{
		Data:   s.Data,
		Types:  s.Types,
		Expiry: s.Expiry,
		Stats:  s.Stats,
	}

	// 3. Создаем файл
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 4. Копируем в Gob и записываем
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(&snap); err != nil {
		return err
	}

	return nil
}

// LoadSnapshot - загружает состояние в файл
func (s *Store) LoadSnapshot(filename string) error {
	// 1. Блокируем для записи
	s.Mu.Lock()
	defer s.Mu.Unlock()

	// 2. Открываем файл
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	// 3. Декодируем из Gob
	var snap SnapshotData
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&snap); err != nil {
		return err
	}

	// 4. Восстанавливаем состояние
	s.Data = snap.Data
	s.Types = snap.Types
	s.Expiry = snap.Expiry
	s.Stats = snap.Stats

	return nil
}
