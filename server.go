package main

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

type Message struct {
	Payload any
}

type R_Payload struct {
	Key string
}

type E_Payload struct {
	EType string
}

type W_Payload struct {
	Key  string
	Data string
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

	messageStatus := make(chan Message)
	messageOk := make(chan struct{})
	messageError := make(chan struct{})
	writingMessage := true

	go func() {
		for {
			var message Message
			if err := server.decode(conn, &message); err != nil {
				if err == io.EOF {
					fmt.Printf("[Server@%s]: Peer:%s disconnected\n", server.listener.Addr().String(), conn.RemoteAddr().String())
				} else {
					fmt.Printf("[Server@%s]: Error while reading from peer..\n", server.listener.Addr().String())
				}
				close(messageStatus)
				return
			}

			switch payload := message.Payload.(type) {

			case W_Payload:
				server.handleWriteFile(conn, payload)

			case R_Payload:
				server.handleReadFile(conn, payload, messageStatus)

			case E_Payload:
				server.handleExecption(conn, payload)
			}
		}
	}()

	go func() {
		for writingMessage {
			select {
			case message := <-messageStatus:
				peerNode, err := server.getConnectionFromConnPool(conn)
				if err != nil {
					fmt.Printf("Peer do not exists..\n")
					close(messageError)
					break
				}

				if err := server.encode(peerNode, &message); err != nil {
					fmt.Printf("[Server@%s]: Decoding error..\n", server.listener.Addr().String())
					close(messageError)
					break
				}

				if _, ok := message.Payload.(E_Payload); ok {
					// fmt.Printf("[Server@%s]: %v\n", server.listener.Addr().String(), (message.Payload.(E_Payload).EType))
					close(messageError)
					break
				}

				close(messageOk)

			case <-messageError:
				writingMessage = false
				break

			case <-messageOk:
				fmt.Printf("[Server@%s]: File served.. successfully...\n", server.listener.Addr().String())
				writingMessage = false
				break
			}
		}
	}()

	select {}
}

func (server *Server) handleWriteFile(conn net.Conn, payload W_Payload) {
	fmt.Printf("[Peer@%s]: Incoming file WRITE request from: %s\n", server.listener.Addr().String(), conn.RemoteAddr().String())

	file, err := server.writeToStorage(payload)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("[Peer@%s]: Successfully written in local storage:\n", server.listener.Addr().String())
	fmt.Printf("[File@%s]: %s\n", server.listener.Addr().String(), string(file))
}

func (server *Server) handleReadFile(conn net.Conn, payload R_Payload, messageStatus chan<- Message) {
	fmt.Printf("[Peer@%s]: Incoming file READ request from: %s for key: '%s'\n", server.listener.Addr().String(), conn.RemoteAddr().String(), payload.Key)

	fileStoragePath := fmt.Sprintf("%s/%s", server.storageLocation, getContentAddress(payload.Key))

	if doesFileExists(fileStoragePath, payload.Key) {
		file, err := server.readFromStorage(payload)
		if err != nil {
			fmt.Printf("Error while reading the file: %s\n ", err)
		}
		fmt.Printf("[Server@%s]: File found, serving file...\n", server.listener.Addr().String())
		fmt.Printf("[File@%s]: %s\n", server.listener.Addr().String(), string(file))

		message := Message{Payload: W_Payload{Key: payload.Key, Data: string(file)}}
		messageStatus <- message

	} else {
		message := Message{Payload: E_Payload{EType: "FILE NOT FOUND"}}
		messageStatus <- message
		fmt.Printf("[Server@%s]: File does not exists.\n", server.listener.Addr().String())
	}
}

func (server *Server) handleExecption(conn net.Conn, payload E_Payload) {
	switch payload.EType {

	case "FILE NOT FOUND":
		fmt.Printf("[Server@%s]: File not found remotely..\n", server.listener.Addr().String())

	default:
		fmt.Printf("[Server@%s]: Some unknown error occured while fetching the file remotely..\n", server.listener.Addr().String())
	}

}
func (server *Server) writeToStorage(payload W_Payload) ([]byte, error) {
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

func (server *Server) readFromStorage(payload R_Payload) ([]byte, error) {
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

func (server *Server) storeFile(payload W_Payload, shareWithPeers bool) {
	payloadSize := reflect.TypeOf(payload).Size()
	fileStoragePath := fmt.Sprintf("%s/%s", server.storageLocation, getContentAddress(payload.Key))

	if doesFileExists(fileStoragePath, payload.Key) {
		file, err := server.readFromStorage(R_Payload{Key: payload.Key})
		if err != nil {
			fmt.Printf("Error while reading the file: %s\n ", err)
		}
		fmt.Printf("[Server@%s]: File already exists.. loading file...\n", server.listener.Addr().String())
		fmt.Printf("[File@%s]: %s\n", server.listener.Addr().String(), string(file))

	} else {
		_, err := server.writeToStorage(payload)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		fmt.Printf("[Server@%s]: Writing %d bytes locally...\n", server.listener.Addr().String(), payloadSize)
	}

	if shareWithPeers {
		if len(server.connectionPool) == 0 {
			fmt.Printf("[Server@%s]: No peer connected, can't share file\n", server.listener.Addr().String())

		} else {
			fmt.Printf("[Server@%s]: Sharing file (%d bytes) with connected peers...\n", server.listener.Addr().String(), payloadSize)

			var wg sync.WaitGroup
			stop := false
			message := Message{Payload: payload}
			for _, peerNode := range server.connectionPool {
				if stop {
					break
				}

				wg.Add(1)
				go func(peerNode net.Conn) {
					defer wg.Done()
					if err := server.encode(peerNode, &message); err != nil {
						fmt.Printf("[Server@%s]: Error while sharing the file with peer..\n", server.listener.Addr().String())
						stop = true
					}
				}(peerNode)
			}

		}
	}

}

func (server *Server) getFile(payload R_Payload) {
	fmt.Printf("[Server@%s]: Requesting file with key: %s\n", server.listener.Addr().String(), payload.Key)
	fileStoragePath := fmt.Sprintf("%s/%s", server.storageLocation, getContentAddress(payload.Key))

	if doesFileExists(fileStoragePath, payload.Key) {
		file, err := server.readFromStorage(payload)
		if err != nil {
			fmt.Printf("Error while reading the file: %s\n ", err)
		}
		fmt.Printf("[Server@%s]: File found locally.. loading file...\n", server.listener.Addr().String())
		fmt.Printf("[File@%s]: %s\n", server.listener.Addr().String(), string(file))

	} else {
		fmt.Printf("[Server@%s]: File does not exists locally..\n", server.listener.Addr().String())
		if len(server.connectionPool) == 0 {
			fmt.Printf("[Server@%s]: No peer connected, can't search for file\n", server.listener.Addr().String())
		} else {
			if err := server.requestFilefromPeers(payload); err != nil {
				fmt.Printf("Error while requesting the file: %s\n", err)
			}
			fmt.Printf("[Server@%s]: File requested from Peers...\n", server.listener.Addr().String())
		}

	}

}

func (server *Server) requestFilefromPeers(payload R_Payload) error {
	message := Message{Payload: payload}
	for _, peer := range server.connectionPool {
		if err := server.encode(peer, &message); err != nil {
			fmt.Printf("[Server@%s]: Error requesting the file from peer..\n", server.listener.Addr().String())
			continue
		}
	}
	return nil
}

func (server *Server) decode(r io.Reader, msg *Message) error {
	return gob.NewDecoder(r).Decode(msg)
}

func (server *Server) encode(r io.Writer, msg *Message) error {
	return gob.NewEncoder(r).Encode(msg)
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
	gob.Register(Message{})
	gob.Register(R_Payload{})
	gob.Register(W_Payload{})
	gob.Register(E_Payload{})
}
