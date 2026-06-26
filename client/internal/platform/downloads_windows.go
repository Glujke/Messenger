//go:build windows

package platform

import "golang.org/x/sys/windows"

func userDownloadsDir() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_Downloads, windows.KF_FLAG_DEFAULT)
}
