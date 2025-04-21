package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// проверяет и создаёт директорию storage в корне проекта
func CheckStorageFolder() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("ошибка получения пути к исполняемому файлу: %w", err)
	}

	dir := filepath.Dir(exePath)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fmt.Errorf("не удалось найти корень проекта `go.mod`")
		}
		dir = parent
	}

	storageDir := filepath.Join(dir, "storage")

	_, err = os.Stat(storageDir)
	if os.IsNotExist(err) {
		err = os.Mkdir(storageDir, 0755)
		if err != nil {
			return fmt.Errorf("ошибка создания директории storage: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("ошибка проверки директории storage: %w", err)
	}

	err = os.Chmod(storageDir, 0755)
	if err != nil {
		return fmt.Errorf("ошибка установки прав для директории storage: %w", err)
	}

	return nil
}
