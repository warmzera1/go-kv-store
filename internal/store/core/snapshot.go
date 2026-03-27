// package core

// import (
// 	"encoding/json"
// 	"os"
// 	"time"
// )

// // SnapshotData - структура для сохранения на диск
// type SnapshotData struct {
// 	Data   map[string]interface{} `json:"data"`
// 	Types  map[string]DataType    `json:"types"`
// 	Expiry map[string]time.Time   `json:"expiry"`
// 	Stats  Stats                  `json:"stats"`
// }

// // SaveSnapshot - сохраняет текущее состояние на диск
// func (s *Store) SaveSnapshot(filename string) error {
// 	// 1. Блокируем хранилище для чтения
// 	s.Mu.Lock()
// 	defer s.Mu.Unlock()

// 	// 2. Собираем данные для сохранения
// 	snap := SnapshotData{
// 		Data:   s.Data,
// 		Types:  s.Types,
// 		Expiry: s.Expiry,
// 		Stats:  s.Stats,
// 	}

// 	// 3. Кодируем в JSON формат
// 	// Создаем читаемый JSON с отступами
// 	data, err := json.MarshalIndent(snap, "", " ")
// 	if err != nil {
// 		return err
// 	}

// 	// 4. Записываем в файл
// 	// 0644 - права доступа: rw-r-r (владелец может читать/писать, остальные только читать)
// 	err = os.WriteFile(filename, data, 0644)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // LoadSnapshot - загружает состояние из файла
// func (s *Store) LoadSnapshot(filename string) error {
// 	// 1. Блокируем хранилище для записи
// 	s.Mu.Lock()
// 	defer s.Mu.Unlock()

// 	// 2. Читаем файл
// 	data, err := os.ReadFile(filename)
// 	if err != nil {
// 		// Если файла нет - пустое состояние
// 		if os.IsNotExist(err) {
// 			return nil
// 		}
// 		return err
// 	}

// 	// 3. Декодируем JSON
// 	var snap SnapshotData
// 	if err := json.Unmarshal(data, &snap); err != nil {
// 		return err
// 	}

// 	// 4. Восстанавливаем состояние
// 	s.Data = snap.Data
// 	s.Types = snap.Types
// 	s.Expiry = snap.Expiry
// 	s.Stats = snap.Stats

// 	return nil
// }

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
