package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DEFAULT_PORT string = ":3000"
	ROOT_STORAGE        = "_STORAGE_"
	PERMISSION          = 0750
)

func main() {
	// fileKey := "KEYFORFILE"
	// path := getContentAddress(fileKey)
	// fileName := getSHAHash(fileKey)
	// if err := createFolder(path); err != nil {
	// 	fmt.Printf("%s\n", err)
	// }
	// if err := createFile(path, fileName); err != nil {
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
	// if err := findFile("/54553/"+path, fileName); err != nil {
	// 	fmt.Println(err)
	// }

	var wg sync.WaitGroup
	wg.Add(2) // Add a counter for the goroutine

	go func() {
		defer wg.Done()
		runServer(":3000", "")
	}()
	time.Sleep(1 * time.Second)
	go func() {
		defer wg.Done()
		runServer(":4000", ":3000")
	}()

	wg.Wait()
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
		return err
	}
	fmt.Printf("[File]:%s\n", string(fileContents))
	return nil
}

func findFile(path string, fileName string) error {
	_, err := os.Stat(getFilePath(ROOT_STORAGE, path, fileName))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist")
		} else {
			fmt.Println("Error:", err)
		}
		return err
	}
	fmt.Println("File exists")
	return nil
}

func runServer(PORT string, p2 string) {
	connectionPool := new(map[string]net.Conn)
	*connectionPool = make(map[string]net.Conn)
	connectionPoolStatus := make(chan string)

	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	defer ln.Close()

	fmt.Printf("[Server@%s]: Server started...\n", PORT[1:])
	if p2 != "" {
		go func() {
			conn, err := net.Dial("tcp", p2)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("[Server@%s]: Connected with %s\n", PORT[1:], conn.RemoteAddr().String())
		}()
	}

	for {
		incomingConnection, err := ln.Accept()
		connectionAddress := incomingConnection.RemoteAddr().String()
		(*connectionPool)[connectionAddress] = incomingConnection
		if err != nil {
			fmt.Printf("%s\n", err)
			continue
		}

		fmt.Printf("[Server@%s]: New peer:%s connected \n", PORT[1:], connectionAddress)

		// fileKey := "KEYFORFILE"
		// path := connectionAddress[6:] + "/" + getContentAddress(fileKey)
		// fileName := getSHAHash(fileKey)
		// if err := createFolder(path); err != nil {
		// 	fmt.Printf("%s\n", err)
		// }
		// if err := createFile(path, fileName); err != nil {
		// 	fmt.Printf("%s\n", err)
		// }

		time.Sleep(time.Second * 1)
		go handleConnection(incomingConnection, PORT, connectionPool, connectionPoolStatus)
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

func handleConnection(conn net.Conn, PORT string, connectionPool *map[string]net.Conn, connectionPoolStatus chan<- string) {
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

func getFilePath(parentDir string, key ...string) string {
	result := ""
	if len(key) == 1 {
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
