package platform

// ClipboardFiles returns local file paths from the system clipboard, if any.
func ClipboardFiles() ([]string, error) {
	return clipboardFiles()
}

// ClipboardImageFile saves a clipboard image to a temp file and returns its path.
func ClipboardImageFile() (path string, ok bool, err error) {
	return clipboardImageFile()
}
