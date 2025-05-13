package app

import "os"

// Проверяет или создает dir для private & public ключей
func CheckKeysFolder(keysDir string) error {
	_, err := os.Stat(keysDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(keysDir, 0770); err != nil {
				return err
			}
		}
		return err
	}
	return nil
}
