package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rudramoorthy/go-chat-app.git/app/config"
	log "github.com/sirupsen/logrus"
)

// Client struct to hold the client details
type Client struct {
	Connection net.Conn
	Message    chan string
	Name       string
	Room       string
}

// Room struct to hold a room details
type Room struct {
	Name    string
	Members map[string]*Client
}

var rooms = make(map[string]*Room)
var m = &sync.Mutex{}

func main() {
	// reading the config file path from the args
	var configfile string
	flag.StringVar(&configfile, "config", "../configs/config.toml", "Configuration file location")
	flag.Parse()
	// loading the toml file
	var configuration = config.LoadConfig(configfile)
	ipaddress := configuration.Server.IP
	port := configuration.Server.Port
	// creating logger
	createLogger(configuration.Server.Logfile)
	// starting the listener with provided ip and port
	listener, err := net.Listen("tcp", ipaddress+":"+port)
	if err != nil {
		log.WithError(err).Fatalln("Error while listening on " + ipaddress + ":" + port)
	}

	defer listener.Close()
	log.Infoln("Server started, tcp Listening on ", listener.Addr())

	// creating a infinite for loop to accept all the incoming requests
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.WithError(err).Error("Error while accepting connection: ")
		}

		go create(connection)
	}

}

// Create both room and client
func create(connection net.Conn) {
	connection.Write([]byte("Please enter room name you want to create/join: "))
	room, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		log.WithError(err).Error("Error while reading room from client: ")
	}
	room = strings.Trim(room, "\r\n")

	connection.Write([]byte("Please enter your name for the room: "))
	name, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		log.WithError(err).Error("Error while reading name from client: ")
	}

	name = validateNameForRoom(strings.Trim(name, "\r\n"), room, connection)
	client := createClient(connection, name)
	createRoom(room, client)

	writeToRoom(room, client, name+" joined..\n")

	go client.sendMessage()
	go client.receiveMessage()
}

// customizing the logrus logger with custom formatter
func createLogger(logfile string) {
	// Create the log file if doesn't exist. And append to it if it already exists.
	f, err := os.OpenFile(logfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	if err != nil {
		// Cannot open log file. Logging to stderr and setting output to stdout
		fmt.Println("Error while creating log file: ", err)
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(f)
	}
	// set at info level for now
	log.SetLevel(log.InfoLevel)
}

// Create client with provided connection and name
func createClient(connection net.Conn, name string) *Client {
	return &Client{
		Connection: connection,
		Name:       name,
		Message:    make(chan string),
		Room:       "",
	}
}

// create room with name and client
func createRoom(name string, client *Client) *Room {
	m.Lock()
	room, ok := rooms[name]
	if ok {
		room.Members[client.Name] = client
	} else {
		room = &Room{
			Name:    name,
			Members: make(map[string]*Client),
		}
		room.Members[client.Name] = client
		rooms[name] = room
	}
	client.Room = name
	m.Unlock()
	return room
}

// write to all the clients in the room except the client who sent the message
func writeToRoom(name string, from *Client, message string) {
	room, ok := rooms[name]
	if ok {
		for _, v := range room.Members {
			if v.Name != from.Name {
				log.Infoln("Sending message to " + v.Name + " from " + from.Name)
				v.Message <- message
			}
		}
	}
}

// validation for unique client name in room
func validateNameForRoom(name string, roomName string, connection net.Conn) string {
	room, ok := rooms[roomName]
	if ok {
		_, ok := room.Members[name]
		if ok {
			for {
				name = requestNewName(name, roomName, connection)
				_, ok := room.Members[name]
				if !ok {
					break
				}
			}
		}
	}

	return name
}

// re-request the name if it is not unique for each room
func requestNewName(clientName string, roomName string, connection net.Conn) string {
	connection.Write([]byte(clientName + " is already taked for the room " + roomName + "\n"))
	connection.Write([]byte("Please enter a different name: "))
	name, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		log.WithError(err).Errorln("Error while reading name from client: ")
	}
	return strings.Trim(name, "\r\n")
}

// send message from client to all other clients in the room
func (client *Client) sendMessage() {
Send:
	for {
		message, err := bufio.NewReader(client.Connection).ReadString('\n')
		if err != nil {
			log.WithError(err).Errorln("Error while reading message from client" + client.Name + " with error: ")
		}
		message = strings.Trim(message, "\r\n")
		if message == ":q" {
			client.leaveRoom()
			break Send
		} else if message == ":cn" {
			client.changeName()
			continue
		} else if message == ":help" {
			client.showHelp()
			continue
		} else {
			message = "<<" + time.Now().Format(time.Stamp) + " (" + client.Name + ")>> : \"" + message + "\""
			writeToRoom(client.Room, client, message)
		}
	}

}

// send message to client in the room
func (client *Client) receiveMessage() {
	for {
		message := <-client.Message
		log.Infoln("Recieved message for client" + client.Name + " in room " + client.Room)
		_, err := client.Connection.Write([]byte(message + "\n"))
		if err != nil {
			log.Infoln("Error while receiving message for client" + client.Name + " in room " + client.Room)
		}
	}

}

// leave room. as of today, leaving the room would quit the connection
// TODO: need to update to allow seemlessly switch the room
func (client *Client) leaveRoom() {
	val, ok := rooms[client.Room]
	if ok {
		m.Lock()
		delete(val.Members, client.Name)
		m.Unlock()
		message := "<<" + time.Now().Format(time.Stamp) + ">>" + client.Name + " left the room...."
		for _, v := range val.Members {
			v.Message <- message
		}
		// closing the connection once the user leaves the room
		client.Connection.Close()
	}
}

// change the name of a particular client in a room
func (client *Client) changeName() {
	oldname := client.Name
	client.Connection.Write([]byte("Please enter new name for the room: "))
	name, err := bufio.NewReader(client.Connection).ReadString('\n')
	if err != nil {
		log.WithError(err).Errorln("Error while reading name from client: ")
	}

	name = validateNameForRoom(strings.Trim(name, "\r\n"), client.Room, client.Connection)
	client.Name = name
	writeToRoom(client.Room, client, oldname+" renamed as "+client.Name)
}

// show options along with client's details
func (client *Client) showHelp() {
	help := "You are currently in the room " + client.Room + " with name " + client.Name +
		"\n options available are, \n :q - to quit the room and connection\n" +
		" :cn - to change the name in room\n"
	client.Connection.Write([]byte(help))
}
