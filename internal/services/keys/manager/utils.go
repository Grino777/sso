package manager

import (
	"fmt"
	"os"
	"sort"
)

func sortPemFiles(pemFiles []os.DirEntry) ([]os.DirEntry, error) {
	sort.Slice(pemFiles, func(i, j int) bool {
		fileIInfo, err := pemFiles[i].Info()
		if err != nil {
			return false
		}
		fileJInfo, err := pemFiles[j].Info()
		if err != nil {
			return false
		}
		return fileIInfo.ModTime().After(fileJInfo.ModTime())
	})

	return pemFiles, nil
}

// Create dirrectory for private & public keys
func createKeysFolder(keysDir string) error {
	const op = opKeysManager + "createKeysFolder"

	if err := os.Mkdir(keysDir, 0700); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Проверяет наличие keys dir
func checkKeysFolder(keysDir string) error {
	const op = opKeysManager + "checkKeysFolder"

	_, err := os.Stat(keysDir)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrDirNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func deleteAllKeys(keysDir string) error {
	const op = opKeysManager + "deleteAllKeys"

	if err := os.RemoveAll(keysDir); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	if err := createKeysFolder(keysDir); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}
