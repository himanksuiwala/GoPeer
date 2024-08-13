package main

// NEED TO LINK THE COONECTION BETWEEN 2 PEERS for FILE TRANSFER
import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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
	cChannel         chan string
	mu               sync.Mutex
	wg               sync.WaitGroup
}

type Payload struct {
	Key  string
	Data string
}

type Message struct {
	Control string
	Payload Payload
}

func instantiatePeer(node string, nodes ...string) *Server {
	return &Server{
		listeningAddress: node,
		peerNodes:        nodes,
		storageLocation:  getStorageParentDir(node),
		connectionPool:   make(map[string]net.Conn),
		cChannel:         make(chan string),
	}
}

func (server *Server) start() {
	var wg sync.WaitGroup

	if err := server.startListening(); err != nil {
		log.Fatalf("Error while spinning up the server! %s\n", err)
	}

	go server.acceptIncomingConnections()

	fmt.Printf("[Server@%s]: Trying to connect with peers, if any...%s\n", server.listener.Addr().String(), strings.Join(server.peerNodes, ", "))
	wg.Add(1)
	go server.connectWithPeers(server.peerNodes, &wg)
	wg.Wait()
}

func (server *Server) startListening() error {
	var err error
	fmt.Printf("Spinning up the peer %s\n", server.listeningAddress)

	server.listener, err = net.Listen("tcp", server.listeningAddress)

	if err != nil {
		return err
	}
	fmt.Printf("[Server@%s]: Server is up & listening...\n", server.listener.Addr().String())
	return nil
}

func (server *Server) acceptIncomingConnections() {
	for {
		conn, err := server.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		}

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		fmt.Printf("[Server@%s]: New incoming peer %s connected \n", server.listener.Addr().String(), conn.RemoteAddr().String())

		server.updateConnectionPool(conn)

		go server.handleIncomingRequest(conn)
	}
}

func (server *Server) handleIncomingRequest(conn net.Conn) {
	// defer func() {
	// 	// conn.Close()
	// 	server.mu.Lock()
	// 	delete(server.connectionPool, conn.RemoteAddr().String())
	// 	server.mu.Unlock()
	// }()

	server.updateConnectionPool(conn)

	messageCh := make(chan Message)

	// Goroutine for reading messages from peers
	go func() {
		for {
			dec := gob.NewDecoder(conn)
			var message Message
			err := dec.Decode(&message)
			if err != nil {
				if err == io.EOF {
					fmt.Printf("[Server@%s]: Peer:%s disconnected\n", server.listener.Addr().String(), conn.RemoteAddr().String())
				} else {
					fmt.Println("Decode error:", err)
				}
				close(messageCh)
				return
			}

			if message.Control == "WRITE" {

				fmt.Printf("[Peer@%s]: Incoming file WRITE request from: %s\n", server.listener.Addr().String(), conn.RemoteAddr().String())
				file, err := server.writeToStorage(message.Payload)
				if err != nil {
					fmt.Printf("%s\n", err)
				}
				fmt.Printf("[Peer@%s]: Successfully written in storage: %s\n", server.listener.Addr().String(), string(file))

			} else if message.Control == "READ" {

				payload := message.Payload
				fmt.Printf("[Peer@%s]: Incoming file READ request from: %s for key: '%s'\n", server.listener.Addr().String(), conn.RemoteAddr().String(), payload.Key)

				fileStoragePath := fmt.Sprintf("%s/%s", server.storageLocation, getContentAddress(payload.Key))

				if doesFileExists(fileStoragePath, payload.Key) {

					file, err := server.readFromStorage(payload)
					if err != nil {
						fmt.Printf("Error while reading the file: %s\n ", err)
					}
					fmt.Printf("[Server@%s]: File found remotely.. serving file...\n", server.listener.Addr().String())
					fmt.Printf("[Server@%s]: %s\n", server.listener.Addr().String(), string(file))

					sentBackPayload := Payload{Key: payload.Key, Data: string(file)}
					message := Message{Control: "WRITE", Payload: sentBackPayload}
					messageCh <- message

				} else {

					fmt.Printf("[Server@%s]: File does not exists remotely\n", server.listener.Addr().String())

				}
				close(messageCh)
			}
		}
	}()

	// Goroutine for writing messages to peers
	go func() {
		for message := range messageCh {
			peerNode, err := server.getConnectionFromConnPool(conn)
			if err != nil {
				fmt.Printf("Peer do not exists..\n")
			}

			enc := gob.NewEncoder(peerNode)
			if err := enc.Encode(message); err != nil {
				fmt.Printf("Error sending message to peer: %s\n", err)
				server.mu.Lock()
				delete(server.connectionPool, peerNode.RemoteAddr().String())
				server.mu.Unlock()

			}
			fmt.Println("message shared -> ", message, peerNode.RemoteAddr().String())
		}
	}()

	select {}
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

func (server *Server) readFromStorage(payload Payload) ([]byte, error) {
	msg, ok := server.validateStorage()
	if !ok {
		return nil, fmt.Errorf(msg)
	}

	fileStoragePath := fmt.Sprintf("%s/%s", server.storageLocation, getContentAddress(payload.Key))
	file, err := readFile(fileStoragePath, payload.Key)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (server *Server) connectWithPeers(peerNodes []string, wg *sync.WaitGroup) {
	for _, peerNode := range peerNodes {
		conn, err := net.Dial("tcp", peerNode)
		if err != nil {
			fmt.Printf("[Server@%s]: Unable to connect with Peer %s, peer is unavailable for connection.. \n", server.listener.Addr().String(), peerNode)
			continue
		}
		fmt.Printf("[Server@%s]: Connection successfully established with  %s\n", server.listener.Addr().String(), conn.RemoteAddr().String())
		server.updateConnectionPool(conn)
		/*
			Required to handle incoming request from other connected peers once the connection is established - handleIncomingRequest.
			Need to find another way
		*/
		go server.handleIncomingRequest(conn)
		break
	}
	wg.Done()
}

func (server *Server) storeFile(payload Payload) {
	payloadSize := reflect.TypeOf(payload).Size()

	if len(server.connectionPool) == 0 {

		_, err := server.writeToStorage(payload)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		fmt.Printf("[Server@%s]: Writing %d bytes to locally...\n", server.listener.Addr().String(), payloadSize)

	} else {

		fmt.Printf("[Server@%s]: Writing %d bytes to connected peers...\n", server.listener.Addr().String(), payloadSize)

		var wg sync.WaitGroup
		stop := false
		message := Message{Control: "WRITE", Payload: payload}
		for _, peerNode := range server.connectionPool {
			if stop {
				break
			}

			wg.Add(1)
			go func(peerNode net.Conn) {
				defer wg.Done()
				enc := gob.NewEncoder(peerNode)
				if err := enc.Encode(message); err != nil {
					fmt.Printf("Error while sharing the file: %s\n", err)
					stop = true
				}
			}(peerNode)
		}

	}
}

func (server *Server) getFile(payload Payload) {
	fmt.Printf("[Server@%s]: Requesting file with key: %s\n", server.listener.Addr().String(), payload.Key)
	fileStoragePath := fmt.Sprintf("%s/%s", server.storageLocation, getContentAddress(payload.Key))

	if len(server.connectionPool) == 0 {

		if doesFileExists(fileStoragePath, payload.Key) {

			file, err := server.readFromStorage(payload)
			if err != nil {
				fmt.Printf("Error while reading the file: %s\n ", err)
			}
			fmt.Printf("[Server@%s]: File found locally.. loading file...\n", server.listener.Addr().String())
			fmt.Printf("[Server@%s]: %s\n", server.listener.Addr().String(), string(file))

		}
		fmt.Printf("[Server@%s]: File does not exists locally..\n", server.listener.Addr().String())

	} else {

		fmt.Printf("[Server@%s]: File does not exists locally..\n", server.listener.Addr().String())
		if err := server.requestFilefromPeers(payload); err != nil {
			fmt.Printf("Error while requesting the file: %s\n", err)
		}
		fmt.Printf("[Server@%s]: File requested from Peers...\n", server.listener.Addr().String())

	}
}

func (server *Server) requestFilefromPeers(payload Payload) error {
	message := Message{Control: "READ", Payload: payload}
	for _, peer := range server.connectionPool {
		enc := gob.NewEncoder(peer)
		if err := enc.Encode(message); err != nil {
			fmt.Printf("Error while transmitting the request: %s\n", err)
			continue
		}
	}
	return nil
}

func (server *Server) sendFiletoPeer() error {
	return nil
}

func (server *Server) updateConnectionPool(conn net.Conn) {
	server.mu.Lock()
	server.connectionPool[conn.RemoteAddr().String()] = conn
	server.mu.Unlock()
}

func (server *Server) getConnectionFromConnPool(conn net.Conn) (net.Conn, error) {
	server.mu.Lock()
	existingNode, isPresent := server.connectionPool[conn.RemoteAddr().String()]
	server.mu.Unlock()
	if isPresent == false {
		return nil, fmt.Errorf("Peer does not exists in connection pool..")
	}
	return existingNode, nil
}

func init() {
	gob.Register(Payload{})
	gob.Register(Message{})
}
