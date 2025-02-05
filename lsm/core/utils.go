package core

import "os"

func EnsurePath(tablePath string) error {
	if _, err := os.Stat(tablePath); err != nil {
		err = os.MkdirAll(tablePath, os.ModePerm)

		if err != nil {
			return os.ErrNotExist
		}
	}

	return nil
}
