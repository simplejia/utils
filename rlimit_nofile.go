// +build !windows

package utils

import (
	"fmt"
	"syscall"
)

func RlimitNofile() string {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("%d", rLimit.Cur)
}
