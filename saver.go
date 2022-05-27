package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

// Create a file named url in directory and save data to it.
func SaveData(data []byte, urlstr string, directory string) error {
	filename := filepath.Join(directory, url.QueryEscape(urlstr))
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("%s: os.OpenFile(): %s", filename, err.Error())
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("f.Write(): %s", err.Error())
	}

	return nil
}
