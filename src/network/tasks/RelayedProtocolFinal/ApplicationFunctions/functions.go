package main

import (
	"bufio"
	Node "chatprotocol/peer"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

var Peer *Node.ChatPeer

func StartNode() {

	// Initialize the node which thus can use get and post functions to manage data. Here by Node we mean every participant like client mod kademlia etc which interacts with the network. It also cobnencts the node with the relay server.
	fmt.Println("Starting Node...")

	relayAddr := "/p2p/relay" // have to build logic about getting Multiple Relays

	//doubt: Will a single node connected to multiple relays or just a single relay.If multiple relays then how to connect the node to them as each connection will give a new peerID as if a new peer is joining.

	var err error
	Peer, err = Node.NewChatPeer(relayAddr)
	if err != nil {
		fmt.Println("Error creating peer:", err)
		return
	}
	ctx := context.Background()

	if err := Peer.Start(ctx); err != nil {
		log.Fatal(err)
	}

}

func GET(targetIP string, targetPort string, route string) (string, error) {

	//reqAddr := targetIP + ":" + targetPort + "/" + route

	reqparams := make(map[string]string)
	parts := strings.Split(route, "/")

	params := strings.Split(parts[1], "&&")

	for i := range len(params) {
		key := strings.Split(params[i], "=")[0]
		value := strings.Split(params[i], "=")[1]

		reqparams[key] = value
	}
	reqparams["Method"] = "GET"
	jsonReq, err := json.Marshal(reqparams)
	if err != nil {
		fmt.Println("[DEBUG]Failed to get req params json")
		return "", err
	}
	ctx := context.Background()
	peerID, err := Peer.ConnectToPeer(ctx, targetIP, targetPort, "testPeer") // doubt: do we need to first run this ConnectToPeer code or can wew directly open a stream.

	if err != nil {
		fmt.Println("Error connecting to node")
	}
	//var peerID peer.ID //doubt how to get peer ID

	// Peer.SendMessage(peerID,string(jsonReq))

	stream, err := Peer.Host.NewStream(ctx, peerID, Node.ChatProtocol)

	if reqparams["peerID"] == "" {
		fmt.Println("Request does not contain peerID, please provide one")
		stream.Write([]byte("Please give peerID"))
		return "", nil
	}

	if err != nil {
		fmt.Println("Error opening a GET stream")
		return "", err
	}
	defer stream.Close()

	_, err = stream.Write([]byte(jsonReq))

	if err != nil {
		fmt.Println("Error writing to GET stream")
		return "", err
	}

	responseReader := bufio.NewReader(stream)
	response, err := responseReader.ReadString('\n')

	if err != nil {
		fmt.Println("Error recieving the Get response")
		return "", err
	}

	fmt.Println(response)

	return response, nil

	//Explaination for logic of streams and how req is served.

	//Every Node (any thing or machine or module that want to use network module) first runs StartNode func which created chat peer ot the node as well as connect it to a relay. Now whenever Get request is given it opens a STREAM at node's end. This stream of sends a req params to the end point mentioned and then listens for respinse at the main stream.
	//At the same time every node has a handleStream func defined with its host that helps in setting serving logics as to how will they see reqParams ans how will they give a response accordingly
	//An additional signalling server is to be used which mainly performs the task of getting peerID of the target node as linp2p relays cant store the data
}

func POST(targetIP string, targetPort string, route string, body map[string]string) (string, error) {
	
	msgJSON, err := json.Marshal(body)
	if err != nil {
		fmt.Println("[DEBUG]Failed to convert to json")
		return "", err
	}
	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	peerID, err := Peer.ConnectToPeer(timeoutCtx, targetIP, targetPort, "testNode2")
	if err != nil {
		fmt.Println("Connection failed:", err)
		return "", err
	}

	stream, err := Peer.Host.NewStream(ctx, peerID, Node.ChatProtocol)
	if err != nil {
		fmt.Println("Error opening a Post stream")
		return "", err
	}
	defer stream.Close()

	_, err = stream.Write([]byte(msgJSON))

	if err != nil {
		fmt.Println("Error writing to POST stream")
		return "", err
	}

	reader := bufio.NewReader(stream)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}
	return response, nil
}
