package storage

import (
	"fmt"
	"strings"
)

type backupInFile struct {
	inMemoryStorage MetricsStorage
	fileStorage     MetricsStorage
}

func NewStorageBackup(inMemoryStorage MetricsStorage, fileStorage MetricsStorage) StorageBackup {
	return &backupInFile{
		inMemoryStorage: inMemoryStorage,
		fileStorage:     fileStorage,
	}
}

func (s *backupInFile) Create() {
	currentState := s.inMemoryStorage.GetMetricValues()

	// выбрать тут формат правильный, чтоб и восстанавливать было легко, и дозаписывать
	// вообще похоже, что нужно в стратегии реализовать оба интерфейса (бекапер и сторадж), и делать всё оттуда
	// либо паркер/форматтер вводить и использовать здесь. Да, наверно так, запихнуть куда нибудь в parser пакет
	sb := strings.Builder{}

	for metricType, metricsByType := range currentState {
		for metricName, metricValue := range metricsByType {
			sb.WriteString(fmt.Sprintf(" %v: %v\n", metricName, metricValue))
		}
	}

	s.fileStorage.Restore(sb.String())
}

func (s *backupInFile) Restore() {

}
