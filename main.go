package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const (
	DEFAULT_PORT string = ":3000"
	ROOT_STORAGE        = "_STORAGE_"
	PERMISSION          = 0750
)

func main() {
	server := instantiatePeer(":3000")
	server2 := instantiatePeer(":4000", ":3000")
	server3 := instantiatePeer(":5000", ":4000", ":3000")
	// server4 := instantiatePeer(":6000", ":5000", ":4000", ":3000")
	// server5 := instantiatePeer(":7000", ":6000", ":5000", ":4000", ":3000")

	server.start()
	time.Sleep(time.Microsecond * 1000)
	// server.getFile()
	server2.start()
	time.Sleep(time.Microsecond * 1000)
	server3.start()
	time.Sleep(time.Microsecond * 1000)
	server3.shareFile()
	// time.Sleep(time.Microsecond * 1000)
	// server3.shareFile()
	// server3.shareFile()
	// time.Sleep(time.Microsecond * 1000)
	// server4.start()
	// server4.shareFile()
	// time.Sleep(time.Microsecond * 1000)
	// server5.start()
	select {}

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
	fmt.Printf("%s\n", "File removed")
	return nil
}

func readFile(path string, fileName string) ([]byte, error) {
	file, err := os.ReadFile(getFilePath(path, fileName))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func findFile(path string, fileName string) error {
	_, err := os.Stat(getFilePath(path, fileName))
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
		n, err := conn.Read(buf) //Reading when connected using telent
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
