package util

import (
	"fmt"
	"os"
)

func Check_delete(filePath string) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Error getting file information:", err)
		return
	}
	if fileInfo.Size() == 0 {
		defer func() {
			err = os.Remove(filePath)
		}()
	}
}
