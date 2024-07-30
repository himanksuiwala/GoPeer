package main

import (
	"fmt"
	"os"
)

func main() {
	createFolder()
	createFile()
	readFile()
	// deleteFile()
}

func createFolder() {
	if err := os.Mkdir("_STORAGE_", 0750); err != nil {
		fmt.Printf("%s\n", err)
	}
}

func createFile() {
	if err := os.WriteFile("_STORAGE_/file1.txt", []byte("This is the first file system implemented from scratch"), 0750); err != nil {
		fmt.Printf("%s\n", err)
	}
}

func deleteFile() {
	if err := os.Remove("_STORAGE_/file1.txt"); err != nil {
		fmt.Printf("%s\n", err)
	}
}

func readFile() {
	fileContents, err := os.ReadFile("_STORAGE_/file1.txt")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Println("[File]:", string(fileContents))
}
