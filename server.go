package main

import (
	"fmt"
	"net"
	"io"
	"bufio"
	"log"
	"strings"
	// "sync"
)

////////////////////////////////////
// const:
////////////////////////////////////
const max_user int = 100

////////////////////////////////////
// struct:
////////////////////////////////////
type User struct {
	uName 				string
	uNick 				string
	passward 			string
	currentChannel		string
	sentMessage chan	Msg
}

type Channel struct {
	channelName 	string
	userArray 		[]string //map[string]bool <- for unique
	getMessage chan Msg
}

type Msg struct {
	uName	string
	says	string
}

type Server struct {
	allUser 	map[string]*User
	allNick		map[string]bool
	allChannel	map[string]*Channel
	uJoin		chan Msg
	uLeft		chan Msg
}

// type chanHolder struct {
// 	setChanMsg chan string
// 	getChanMsg chan string
// }

////////////////////////////////////
// Global Variable:
////////////////////////////////////

// // Create map of unique user name:
// var userMap = map[string]*User{}

// // Create map of unique nick name: (only for check)
// var nickMap = map[string]bool{}

// // Create map of unique channel name as key and string array as array:
// var channelMap = map[string]*Channel{}

////////////////////////////////////
// functions:
////////////////////////////////////

// NewChanHolder returns a new Holder backed by Channels.
// func NewChanHolder() Holder {
// 	h := chanHolder {
// 	  setValCh: make(chan string),
// 	  getValCh: make(chan string),
// 	}
// 	go h.mux()
// 	return h
// }

// func (h chanHolder) mux() {
// 	var value string
// 	for {
// 	  select {
// 	  case value = <-h.setValCh: // set the current value.
// 	  case h.getValCh <- value: // send the current value.
// 	  }
// 	}
//   }

//   func (h chanHolder) Get() string {
// 	return <-h.getValCh
//   }

//   func (h chanHolder) Set(s string) {
// 	h.setValCh <- s
//   }
  //-----------------------------------//

func only_take_isprint(s string) string {
	var s_return = ""
	var len int = len(s)

	for i := 0; i < len; i++ {
		if s[i] >= 32 && s[i] <= 126 {
			s_return += string(s[i]) // s[i] is char, so need to use string() to conver type
		}
	}
	return s_return
}

func chatFormat(u User) string {
	return "\n"  + u.currentChannel + "/" + u.uNick + " > "
}



func handleConnection(conn net.Conn, server Server) {
	defer conn.Close()

	// 1. INPUT user name:
	io.WriteString(conn, "Please enter your name: ")
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	uname := only_take_isprint(scanner.Text())
	// if user name already exist:
	// 		- check password
	//		- sent user into chat room
	if existUser, ok := server.allUser[uname]; ok {
		io.WriteString(conn, "Please enter your password: ")
		scanner.Scan()
		inputPassword := scanner.Text()
		for i := 0; (inputPassword != existUser.passward || i > 3); i++ {
			io.WriteString(conn, "Try again: ")
			scanner.Scan()
			inputPassword = scanner.Text()
		}
	}else{
		// if it's NEW user:
		//		- add user name into userMap
		//		- add uName
		server.allUser[uname] = new(User)
		server.allUser[uname].uName = uname
		// 2. INPUT nick name:
		io.WriteString(conn, "Please enter your nick name: ")
		scanner.Scan()
		nickname := scanner.Text()
		// -> use nickMap to check if the nickName has been taken:
		_, nickOk := server.allNick[nickname]
		for nickOk {
			io.WriteString(conn, "This nickName has been taken, please choose another nick name: ")
			scanner.Scan()
			nickname := scanner.Text()
			_, nickOk = server.allNick[nickname]
		}
		server.allNick[nickname] = true
		server.allUser[uname].uNick = nickname
		// 3. INPUT password:
		io.WriteString(conn, "Please enter your password: ")
		scanner.Scan()
		inputPassword := scanner.Text()
		server.allUser[uname].passward = inputPassword
		fmt.Println("userMap = \n", server.allUser) // for debug...
		fmt.Println("nickMap = \n", server.allNick) // for debug...
	}
	// 4. ENTER chat room:
	io.WriteString(conn, "Please enter chat room: ")
	scanner.Scan()
	chatroom := scanner.Text()
	// If the chat room is exist:
	if chatRoom, ok := server.allChannel[chatroom]; ok {
		// update user info of currentChannel
		server.allUser[uname].currentChannel = chatroom
		// add user into channel (Q: do I need to check user first?! Or use map struct)
		chatRoom.userArray = append(chatRoom.userArray, uname)
	// If the chat room is not exist:
	} else {
		server.allChannel[chatroom] = new(Channel)
		server.allChannel[chatroom].channelName = chatroom
		server.allChannel[chatroom].userArray = append(server.allChannel[chatroom].userArray, uname)
		server.allChannel[chatroom].getMessage = make(chan Msg, 10)
		// update user info of currentChannel
		server.allUser[uname].currentChannel = chatroom
		server.allUser[uname].sentMessage = make(chan Msg, 10)
	}
	// for debug...
	for k,v := range server.allChannel {
		fmt.Print("\nChannel: " + k + " , userArray = ")
		fmt.Println(v.userArray)
	}

	io.WriteString(conn, "\n" + uname + " join the " + chatroom + "\n")

	// // Wait user to input: -> maybe create a real function... and use gorutine
	func (u string) {
		io.WriteString(conn, chatFormat(*server.allUser[u]))
		for scanner.Scan() {
			// get input message and tolower:
			commandSplit := strings.SplitN(scanner.Text(), " ", 2)
			// use regex to distinct COMMAND or chat message:
			// user's current chatroom
			// userCurrChatRoomName := userMap[u].currentChannel

			switch strings.ToLower(commandSplit[0]) {
			case "/exit":
				break
			case "/nick": // allNick		map[string]bool
				if len(commandSplit) != 2 {
					fmt.Println("Usage: /nick [new nick name]")
				}else{
					// Change user's nick name:
					// 1. modify server.allUser.uNick
					// 2. delete server.allNick: delete(m, "route")
					// 3. add new nickName: server.allNick
					oriNick := server.allUser.uNick
				}
			case "/join":
			case "/names": // allUser 	map[string]*User
			case "/list": // allChannel	map[string]*Channel
			case "/privmsg":
			//case "/pass_nick_user":
			default:
			}
			io.WriteString(conn, chatFormat(*server.allUser[u]))
		}
	}(uname)
}

////////////////////////////////////
// MAIN:
// 		check the listening port: lsof -nP +c 15 | grep LISTEN
////////////////////////////////////
func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:9000")
	defer ln.Close()
	if err != nil {
		// handle error
		log.Fatalln(err.Error())
	}

	// Create Server struct:
	serverStruct := Server{
		allUser:	make(map[string]*User),
		allNick:	make(map[string]bool),
		allChannel: make(map[string]*Channel),
		uJoin:	make(chan Msg),
		uLeft:	make(chan Msg),
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Fatalln(err.Error())
		}
		go handleConnection(conn, serverStruct)
		defer conn.Close()
	}
}
