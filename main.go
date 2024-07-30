package main

import (
	"fmt"
	"net"
	"os"
)

const PORT string = ":3000"

func main() {
	// createFolder()
	// createFile()
	// readFile()
	// deleteFile()

	runServer()
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

func runServer() {
	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	defer ln.Close()

	fmt.Printf("Server started on Port %s\n", PORT)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		fmt.Printf("[Server@%s]: %s\n", PORT, conn.RemoteAddr().String())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	fmt.Printf("[Server@%s]: %s\n", PORT, string(buf[:n]))
}
