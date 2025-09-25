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

func CloneMap[K comparable, V any](m map[K]V) map[K]V {
	clone := make(map[K]V, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}
