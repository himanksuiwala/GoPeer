package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	PORT         string = ":3000"
	ROOT_STORAGE        = "_STORAGE_"
	PERMISSION          = 0750
)

func main() {
	// fileKey := "KEYFORFILE"
	// path := getContentAddress(fileKey)

	// if err := createFolder(path); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }
	// if err := createFile(path, "firstFileUsingGo.txt"); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }
	// if err := readFile(path, "firstFileUsingGo.txt"); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }
	// if err := deleteFile(path, "firstFileUsingGo.txt"); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }
	// if err := removeFolder(ROOT_STORAGE); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }
	// if err := runServer(); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }
}

func createFolder(path string) error {
	if err := os.MkdirAll(getFilePath(ROOT_STORAGE, path), PERMISSION); err != nil {
		return err
	}
	fmt.Printf("%s\n", "Folder created")
	return nil
}

func removeFolder(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	fmt.Printf("%s\n", "Folder removed")
	return nil
}

func createFile(path string, fileName string) error {
	if err := os.WriteFile(getFilePath(ROOT_STORAGE, path, fileName), []byte("This is the first file system implemented from scratch"), PERMISSION); err != nil {
		return err
	}
	fmt.Printf("%s\n", "File created")
	return nil
}

func deleteFile(path string, fileName string) error {
	if err := os.Remove(getFilePath(ROOT_STORAGE, path, fileName)); err != nil {
		return err
	}
	fmt.Printf("%s\n", "File removed")
	return nil
}

func readFile(path string, fileName string) error {
	fileContents, err := os.ReadFile(getFilePath(ROOT_STORAGE, path, fileName))
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("[File]:%s\n", string(fileContents))
	return nil
}

func runServer() error {
	connectionPool := new(map[string]net.Conn)
	*connectionPool = make(map[string]net.Conn)
	connectionPoolStatus := make(chan string)

	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	defer ln.Close()

	fmt.Printf("[Server@%s]: Server started...\n", PORT[1:])

	for {
		incomingConnection, err := ln.Accept()
		connectionAddress := incomingConnection.RemoteAddr().String()
		(*connectionPool)[connectionAddress] = incomingConnection
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		fmt.Printf("[Server@%s]: New peer connected %s\n", PORT[1:], connectionAddress)

		go handleConnection(incomingConnection, connectionPool, connectionPoolStatus)
		go func() {
			for err := range connectionPoolStatus {
				if err == "ZERO_PEERS" {
					fmt.Printf("[Server@%s]: Terminating the server...\n", PORT[1:])
					panic("0 peers found, server closed")
				}
			}
		}()
	}
}

func handleConnection(conn net.Conn, connectionPool *map[string]net.Conn, connectionPoolStatus chan<- string) {
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		activeConnection := conn.RemoteAddr().String()
		if err == io.EOF {
			fmt.Printf("[Server@%s]: Peer[%s] disconnected\n", PORT[1:], conn.RemoteAddr().String())

			delete(*connectionPool, activeConnection)

			if len(*connectionPool) == 0 {
				connectionPoolStatus <- "ZERO_PEERS"
			}
			break
		}
		if err != nil {
			fmt.Printf("%s\n", err)
			delete(*connectionPool, activeConnection)
			break
		}

		msg := string(buf[:n])

		fmt.Printf("[Peer@%s]: %s\n", conn.RemoteAddr().String(), msg)
	}
}

func getContentAddress(key string) string {
	hash := sha1.Sum([]byte(key))
	hashInString := hex.EncodeToString(hash[:])
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

func getFilePath(parentDir string, key ...string) string {
	result := ""
	if len(key) == 1 {
		result = fmt.Sprintf("%s/%s", parentDir, key[0])
	} else if len(key) == 2 {
		result = fmt.Sprintf("%s/%s/%s ", parentDir, key[0], key[1])
	}
	return result
}
