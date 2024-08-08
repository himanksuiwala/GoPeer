package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
)

type Server struct {
	listeningAddress string
	peerNodes        []string
	storageLocation  string
	listener         net.Listener
	connectionPool   map[string]net.Conn
}

type Payload struct {
	Key  string
	Data string
}

func instantiatePeer(node string, nodes ...string) *Server {
	return &Server{
		listeningAddress: node,
		peerNodes:        nodes,
		storageLocation:  getStorageParentDir(node),
		connectionPool:   make(map[string]net.Conn),
	}
}

func (server *Server) start() {
	fmt.Printf("Spinning up the peer %s\n", server.listeningAddress)

	msg, ifExists := server.validatePeerStorage()
	if ifExists != true {
		fmt.Printf("[Server@%s]: %s\n", server.listeningAddress[1:], msg)
		fmt.Printf("[Server@%s]: Please check the path for storage and try again... still you'll be listening for peers\n", server.listeningAddress[1:])
	} else {
		fmt.Printf("[Server@%s]: Storage location validated\n", server.listeningAddress[1:])
	}

	if err := server.startListening(); err != nil {
		fmt.Printf("%s\n", err)
	}
	go server.acceptIncomingConnections()

	fmt.Println("[Server@"+server.listeningAddress[1:]+"]: Trying to connect with peers, if any...", strings.Join(server.peerNodes, ", "))
	var wg sync.WaitGroup
	wg.Add(1)
	go server.connectWithPeers(server.peerNodes, &wg)
	wg.Wait()
}

func (server *Server) startListening() error {
	var err error
	server.listener, err = net.Listen("tcp", server.listeningAddress)

	if err != nil {
		return err
	}
	fmt.Printf("[Server@%s]: Server is up & listening...\n", server.listeningAddress[1:])
	return nil
}

func (server *Server) acceptIncomingConnections() {
	for {
		conn, err := server.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		fmt.Printf("[Server@%s]: New peer%s connected \n", server.listeningAddress[1:], conn.RemoteAddr().String()[5:])
		go server.handleIncomingRequest(conn)
	}
}

func (server *Server) handleIncomingRequest(conn net.Conn) {
	defer conn.Close()
	for {
		var payload Payload

		dec := gob.NewDecoder(conn)

		err := dec.Decode(&payload)
		if err != nil {
			log.Println("Decode error:", err)
			return
		}

		if err == io.EOF {
			fmt.Printf("[Server@%s]: Peer:%s disconnected\n", server.listeningAddress, conn.RemoteAddr().String()[5:])
			break
		} else if err != nil {
			fmt.Printf("[Server@%s]: Nothing to readmore, closing the connection...\n", server.listeningAddress)
			conn.Close()
		}

		file, err := server.writeToStorage(payload)
		if err != nil {
			fmt.Printf("%s\n", err)
		}

		fmt.Printf("[Peer@%s]: Succesfully written in storage: %s\n", conn.RemoteAddr().String()[5:], string(file))
	}
}

func (server *Server) writeToStorage(payload Payload) ([]byte, error) {
	fileStoragePath := fmt.Sprintf("%s/%s", server.storageLocation, getContentAddress(payload.Key))

	if err := createFolder(fileStoragePath); err != nil {
		fmt.Printf("%s\n", err)
		return nil, err
	}
	if err := createFile(fileStoragePath, payload.Key, []byte(string(payload.Data))); err != nil {
		fmt.Printf("%s\n", err)
		return nil, err
	}
	file, err := readFile(fileStoragePath, payload.Key)
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil, err
	}

	return file, nil
}

func (server *Server) readFromStorage() error {
	return fmt.Errorf("err")
}

func (server *Server) connectWithPeers(peerNodes []string, wg *sync.WaitGroup) {
	for _, peerNode := range peerNodes {
		conn, err := net.Dial("tcp", peerNode)
		if err != nil {
			fmt.Printf("[Server@%s]: Unable to connect with Peer %s, peer is unavailable for connection.. \n", server.listeningAddress[1:], peerNode)
			continue
		}
		fmt.Printf("[Server@%s]: Connected with Peer%s\n", server.listeningAddress[1:], conn.RemoteAddr().String()[9:])
		server.connectionPool[peerNode] = conn
		fmt.Printf("[Server@%s]: Connection Pool...", server.listeningAddress[1:])
		for _, val := range server.connectionPool {
			fmt.Printf("%s, ", val.RemoteAddr().String())
		}
		fmt.Println("")
	}
	wg.Done()
}

func (server *Server) validatePeerStorage() (string, bool) {
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

func (server *Server) shareFile() {
	payload := Payload{Key: "ThisismyKEY", Data: "This data is sent over the network."}
	payloadSize := reflect.TypeOf(payload).Size()

	fmt.Printf("[Server@%s]: Writing %d bytes to connected peers...\n", server.listeningAddress[1:], payloadSize)

	for _, peerNode := range server.connectionPool {
		enc := gob.NewEncoder(peerNode)

		if err := enc.Encode(payload); err != nil {
			fmt.Printf("%s\n", err)
			continue
		}
	}
}

func (server *Server) getFile() {
	if err := server.readFromStorage(); err != nil {
		fmt.Printf("[Server@%s]: File does not exists in locally..\n", server.listeningAddress[1:])
	}

	if err := server.getFilefromPeers(); err != nil {
		fmt.Printf("%s\n", err)
	}
}

func (server *Server) getFilefromPeers() error {
	for _, peer := range server.connectionPool {
		//Do something to get files from peer
		fmt.Println(peer)
	}
	return nil
}
func (server *Server) sendFiletoPeer() error {
	return nil
}

func getStorageParentDir(node string) string {
	return fmt.Sprintf("_STORAGE_@%s", node[1:])
}
