package providers

import "os"

func OnDisk(path string, size int64) bool {
	if size <= 0 {
		return false
	}
	fi, err := os.Stat(path)
	return err == nil && fi.Size() == size
}
