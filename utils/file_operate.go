package utils

import "os"

// IsExistLocalFile @description: 判断一个本地文件或文件夹是否存在
// @parameter path(文件路径)
// @return bool(true:存在；false:不存在)
func IsExistLocalFile(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
