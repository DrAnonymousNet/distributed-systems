package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)



type client_channel chan <- string

type clientInfo struct {
	name string
	client_channel chan <- string
}

type privateMessage struct {
	sender string
	message string
	target string
}


var (
	entering = make(chan clientInfo)
	leaving  = make(chan clientInfo)
	message = make(chan string)
	private = make(chan privateMessage)
	clients = make(map[client_channel]clientInfo)
)


func main(){
	listener, err := net.Listen("tcp", "localhost:8000")
	if err!= nil {
		log.Fatal(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleConn(conn)
		}
	
}



func broadcaster(){
	for {
		select{
		case msg := <-message:
			broadcast(msg)
		case pm := <- private:
			sendPrivateMessage(pm)

		case client := <-entering:
			clients[client.client_channel] = client

		case client := <- leaving:
			delete(clients, client.client_channel)
			close(client.client_channel)
		}
	}

}



func handleConn(conn net.Conn){
	cli := make(chan string)
	me := conn.RemoteAddr().String()

	clientInfo := clientInfo{
		name: me,
		client_channel: cli,
	}

	go clientWriter(conn, cli)


	cli <-"you are " + clientInfo.name
	entering <- clientInfo
	message <- me + " has entered the chat"

	input := bufio.NewScanner(conn)
	//for {

	
	for input.Scan(){
		text := input.Text()

		if strings.HasPrefix(text, "/private"){

			
			cmdParts := strings.Fields(text)
			if len(cmdParts) >= 3 {
				targetClient := cmdParts[1]
				text = strings.Join(cmdParts[2:], " ")
				private <- privateMessage{
					sender: me,
					message: strings.Join(strings.Fields(text), " "),
					target: targetClient,
				}
				}else {
				message <- "Usage: /private <user> <message>"
			}
		}else if strings.HasPrefix(text, "/quit"){
			leaving <- clientInfo
			message <- me + " has left the chat"
			conn.Close()
			break

		}else{
			message <- me + ": " + input.Text()
	}
//}
	// leaving <- clientInfo
	// message <- me + " has left the chat"
	// conn.Close()
}
}

func findClient(target string) (clientInfo, bool){
	for _, info := range clients{
		if info.name == target{
			return info, true
		}
	}
	return clientInfo{}, false
}

func broadcast(msg string){
	for _, client := range clients{
		client.client_channel <- msg
	}
}


func clientWriter(conn net.Conn, cli <- chan string){
	for msg := range cli{
		fmt.Fprintln(conn, msg)
	}
}

func sendPrivateMessage(pm privateMessage) {
	targetClient, ok := findClient(pm.target)
	if ok {
		targetClient.client_channel <- pm.sender + " (private): " + pm.message
	} else {
		message <- "User " + pm.target + " not found or offline."
	}
}