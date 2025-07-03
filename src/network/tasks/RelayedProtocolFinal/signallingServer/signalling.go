package signalling

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
)

//This server is used primarily to give the peerIDs or requested node so complete address of peers can be made for libp2p to est connections.

var ReqFormat struct {
	Type string `json:"type"`
	Addr  string `json:"ip"`
}

var IDmap map[string]peer.ID

func StartServer(port string) {
	//starts a signalling server and esatblishes it for serving the requests

	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server listening on port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleRequests(conn)
	}
}

func handleRequests(conn net.Conn) {
	defer conn.Close()
	fmt.Println("New client connected:", conn.RemoteAddr())

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err)
		return
	}

	fmt.Println("Received:", string(buf[:n]))

	// Unmarshal JSON data into ReqFormat struct
	var req = ReqFormat
	err = json.Unmarshal(buf[:n], &req)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	// fmt.Printf("Parsed request: %+v\n", req)

	if req.Type ==  "register" {
	var ip string
	var peerID peer.ID
	// making vars for now, will process addr given accordinfg to structuree later
	IDmap[ip] = peerID
	conn.Write([]byte("registered"))
	}

	if req.Type == "getID" {

		_,err = conn.Write([]byte(IDmap[req.Addr]))

		if err!=nil {
			fmt.Println("error sending ID to the node")
		}

	}

}
