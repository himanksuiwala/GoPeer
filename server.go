package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Server struct {
	listeningAddress string
	peerNodes        []string
	storageLocation  string
	listener         net.Listener
}

func instantiatePeer(node string, nodes ...string) *Server {
	return &Server{
		listeningAddress: node,
		peerNodes:        nodes,
		storageLocation:  getStorageParentDir(node),
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
	go server.acceptConnections()

	if len(server.peerNodes) > 0 {
		fmt.Println("[Server@"+server.listeningAddress[1:]+"]: Trying to connect with peers...", strings.Join(server.peerNodes, ", "))
		go server.connectWithPeers(server.peerNodes)
	}
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

func (server *Server) acceptConnections() {
	for {
		conn, err := server.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		fmt.Printf("[Server@%s]: New peer%s connected \n", server.listeningAddress[1:], conn.RemoteAddr().String()[5:])
		go server.handleConnections(conn)
	}
}

func (server *Server) handleConnections(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			fmt.Printf("[Server@%s]: Peer:%s disconnected\n", server.listeningAddress, conn.RemoteAddr().String()[5:])
			break
		}
		if err != nil {
			fmt.Printf("[Server@%s]: Nothing to readmore, closing the connection...\n", server.listeningAddress)
			conn.Close()
		}

		msg := string(buf[:n])
		fmt.Printf("[Peer@%s]: %s\n", conn.RemoteAddr().String()[5:], msg)
	}
}

func (server *Server) connectWithPeers(peerNodes []string) {
	for _, peerNode := range peerNodes {
		conn, err := net.Dial("tcp", peerNode)
		if err != nil {
			fmt.Printf("[Server@%s]: Unable to connect with Peer %s, peer is unavailable for connection.. \n", server.listeningAddress[1:], peerNode)
			continue
		}
		fmt.Printf("[Server@%s]: Connected with Peer%s\n", server.listeningAddress[1:], conn.RemoteAddr().String()[9:])
	}
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

func getStorageParentDir(node string) string {
	return fmt.Sprintf("_STORAGE_@%s", node[1:])
}
