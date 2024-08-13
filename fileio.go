package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	DEFAULT_PORT string = ":3000"
	ROOT_STORAGE        = "_STORAGE_"
	PERMISSION          = 0750
)

func getContentAddress(key string) string {
	hashInString := getSHAHash(key)
	hashLen := len(hashInString)
	blockSize := 5
	slices := hashLen / blockSize
	paths := make([]string, slices)
	for i := 0; i < slices; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashInString[from:to]
	}

	return strings.Join(paths, "/")
}

func (server *Server) validateStorage() (string, bool) {
	path := server.storageLocation
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "Directory does not exist.", false
	} else if err != nil {
		return "Unknown error occured:", false
	} else if !info.IsDir() {
		return "Path exists, but it is not a directory.", false
	}
	return "Directory exists.", true
}

func createFolder(path string) error {
	if err := os.MkdirAll(getFilePath(path), PERMISSION); err != nil {
		return err
	}
	return nil
}

func removeFolder(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func createFile(path string, fileName string, fileData []byte) error {
	if err := os.WriteFile(getFilePath(path, fileName), []byte(fileData), PERMISSION); err != nil {
		return err
	}
	return nil
}

func deleteFile(path string, fileName string) error {
	if err := os.Remove(getFilePath(ROOT_STORAGE, path, fileName)); err != nil {
		return err
	}
	return nil
}

func readFile(path string, fileName string) ([]byte, error) {
	file, err := os.ReadFile(getFilePath(path, fileName))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func doesFileExists(path string, fileName string) bool {
	_, err := os.Stat(getFilePath(path, fileName))
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func getFilePath(parentDir string, key ...string) string {
	result := ""
	if len(key) == 0 {
		result = fmt.Sprintf("%s", parentDir)
	} else if len(key) == 1 {
		result = fmt.Sprintf("%s/%s", parentDir, key[0])
	} else if len(key) == 2 {
		result = fmt.Sprintf("%s/%s/%s ", parentDir, key[0], key[1])
	} else if len(key) == 3 {
		result = fmt.Sprintf("%s/%s/%s/%s ", parentDir, key[0], key[1], key[2])
	}
	return result
}

func getSHAHash(key string) string {
	hash := sha1.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

func getStorageParentDir(node string) string {
	parsed := regexp.MustCompile("[:.]").ReplaceAll([]byte(node), []byte("_"))
	return fmt.Sprintf("STORAGE@%s", string(parsed))
}
