//go:build !linux

package loggeradapter

import "os"

func chown(_ string, _ os.FileInfo) error {
	return nil
}
