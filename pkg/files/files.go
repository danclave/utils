package fileutil

import (
	"os"
	"path/filepath"
)

// Read reads the entire contents of a file and returns it as a string.
func Read(path string) (string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Write writes the given content to the file, overwriting if it exists.
func Write(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// Exists checks whether a file or directory exists at the given path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Append adds content to the end of the file, creating it if necessary.
func Append(path, content string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	return err
}

func IsMarkdown(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && filepath.Ext(path) == ".md"
}
