package sys

import "runtime"

// 是否调试模式
func IsDebug() bool {
	return _globalConfig.Debug
}

// 是否灰度发布模式
func IsHdRelease() bool {
	return _globalConfig.HdRelease
}

func IsLinux() bool {
	return runtime.GOOS == "linux"
}