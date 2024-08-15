package main

import (
	"time"
)

func main() {
	server := instantiatePeer("127.0.0.1:3001")
	server2 := instantiatePeer("127.0.0.1:4000", "127.0.0.1:3001")

	go func() { server.start() }()

	time.Sleep(time.Second * 2)

	go func() { server2.start() }()

	/*NO INCOMING REQUEST on :4000 to get data
	connect with peer -> handle incomring request, connection pool
	*/
	// w_payload := W_Payload{Key: "SOMEID_18", Data: "THIS IS WRITTEN from :4000 to :3000 --- 9"}
	r_payload := R_Payload{Key: "SOMEID_8"}

	time.Sleep(time.Second * 2)
	server.getFile(r_payload)
	// server2.storeFile(Payload{Key: "SOMEID_17", Data: "THIS IS WRITTEN from :4000 to :3000 --- 9"})
	// server2.storeFile(Payload{Key: "SOMEID_18", Data: "THIS IS WRITTEN from :4000 to :3000 --- 8"})

	select {}

}
