package main

import (
	"fmt"
	"net"
	"io"
	"bufio"
	"log"
	"strings"
)

////////////////////////////////////
// struct:
////////////////////////////////////
type User struct {
	uName 				string
	uNick 				string
	passward 			string
	currentChannel		string
	conn				net.Conn
}

type Channel struct {
	channelName 	string
	userMap 		map[string]*User
}

type Msg struct {
	uName	string
	says	string
}

type Server struct {
	allUser 	map[string]*User
	allNick		map[string]*User
	allChannel	map[string]*Channel
}

// type chanHolder struct {
// 	setChanMsg chan Msg
// 	getChanMsg chan Msg
// }

////////////////////////////////////
// functions:
////////////////////////////////////

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

func private_msg(destConn net.Conn, msg Msg) {
	io.WriteString(destConn, "\n" + msg.uName + " sent you message: " + msg.says)
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
		server.allNick[nickname] = server.allUser[uname]
		server.allUser[uname].uNick = nickname
		// 3. INPUT password:
		io.WriteString(conn, "Please enter your password: ")
		scanner.Scan()
		inputPassword := scanner.Text()
		server.allUser[uname].passward = inputPassword
		// fmt.Println("userMap = \n", server.allUser) // for debug...
		// fmt.Println("nickMap = \n", server.allNick) // for debug...
		// 4. Initionalize the message:
		server.allUser[uname].conn = conn
	}
	// 4. ENTER chat room:
	io.WriteString(conn, "Please enter chat room: ")
	scanner.Scan()
	chatroom := scanner.Text()
	// If the chat room is exist:
	if chatRoom, ok := server.allChannel[chatroom]; ok {
		// update user info of currentChannel
		server.allUser[uname].currentChannel = chatroom
		// add user into channel:

		chatRoom.userMap[uname] = server.allUser[uname]
	// If the chat room is not exist:
	} else {
		server.allChannel[chatroom] = new(Channel)
		server.allChannel[chatroom].channelName = chatroom
		server.allChannel[chatroom].userMap = make(map[string]*User)
		server.allChannel[chatroom].userMap[uname] = server.allUser[uname]
		// update user info of currentChannel
		server.allUser[uname].currentChannel = chatroom
	}
	// for debug...
	// for k,v := range server.allChannel {
	// 	fmt.Print("\nChannel: " + k + " , userArray = ")
	// 	fmt.Println(v.userMap)
	// }

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
			/////////// EXIT:
			case "/exit":
				// 1. delete user in the channel
				// 2. user's currentChannel set to ""
				thisChannel := server.allUser[u].currentChannel
				delete(server.allChannel[thisChannel].userMap, u)
				server.allUser[u].currentChannel = ""
				return
			/////////// NICK:
			case "/nick": // allNick map[string]bool
				if len(commandSplit) != 2 {
					io.WriteString(conn, "Usage: /nick newNickName")
				}else{
					// Change user's nick name:
					// 1. modify server.allUser.uNick
					// 2. delete server.allNick: delete(m, "route")
					// 3. add new nickName: server.allNick
					oriNick := server.allUser[u].uNick
					newNick := strings.Trim(commandSplit[1], " ")
					server.allUser[u].uNick = newNick
					delete(server.allNick, oriNick)
					server.allNick[newNick] = server.allUser[uname]
				}
			/////////// JOIN:
			case "/join":
				if len(commandSplit) != 2 {
					io.WriteString(conn, "Usage: /join anotherChannel")
				}else{
					// Let user out of channel and join another one
					// 1. Find the original channel an delete the user
					oriChannel := server.allUser[u].currentChannel
					newChannel := strings.Trim(commandSplit[1], " ")
					delete(server.allChannel[oriChannel].userMap, u)
					// 2. Check new channel: if exist add user, if not create one
					if ch, ok := server.allChannel[newChannel]; ok{
						ch.userMap[u] = server.allUser[u]
					}else{
						// create new channel:
						server.allChannel[newChannel] = new(Channel)
						server.allChannel[newChannel].channelName = newChannel
						server.allChannel[newChannel].userMap = make(map[string]*User)
						server.allChannel[newChannel].userMap[u] = server.allUser[u]
					}
					// 3. server.allUser[u].currentChannel => new channel
					server.allUser[u].currentChannel = newChannel
				}
			/////////// NAMES: List all nick name
			case "/names": // allNick map[string]*User
				if len(commandSplit) == 2 {
					// finding specific channel and list all user:
					for room, chStruct := range server.allChannel {
						if room == commandSplit[1] {
							io.WriteString(conn, "Channel - " + room + "\n\tUsers: ")
							for username, _ := range chStruct.userMap {
								io.WriteString(conn, username + " ")
							}
						}else{
							io.WriteString(conn, "Cannot find this channel.\n")
						}
					}
				}else if len(commandSplit) == 1 {
					// List all user name: seprate by channel
					for room, chStruct := range server.allChannel {
						io.WriteString(conn, "Channel - " + room + "\n\tUsers: ")
						for username, _ := range chStruct.userMap {
							io.WriteString(conn, username + " ")
						}
						io.WriteString(conn, "\n")
					}
				}else{
					io.WriteString(conn, "Usage: /names [channel name]")
				}
			/////////// LIST:
			case "/list": // allChannel map[string]*Channel
				if len(commandSplit) != 1 {
					io.WriteString(conn, "Usage: /list")
				}else{
					// List all Channel:
					for room, _ := range server.allChannel {
						io.WriteString(conn, "Channel - " + room + "\n")
					}
					io.WriteString(conn, "\n")
				}
			/////////// PRIVMSG:
			case "/privmsg":
				privArgv := strings.SplitN(scanner.Text(), " ", 3)
				if len(privArgv) != 3 {
					io.WriteString(conn, "Usage: /privmsg NickName Message")
				}else{
					// Get this user's nick name:
					myNick := server.allUser[u].uNick
					// Get dst conn:
					sentToUser :=  privArgv[1]
					if usrPtr, ok := server.allNick[sentToUser]; ok{
						dstConn := usrPtr.conn
						text := privArgv[2]
						thisMsg := Msg{
							uName: myNick,
							says: text,
						}
						private_msg(dstConn, thisMsg)
					}else{
						io.WriteString(conn, "Sorry! I can't find this user!\nPlease try again!")
					}
					
				}
			/////////// PART:
			case "/part":
				if len(commandSplit) != 1 {
					io.WriteString(conn, "Usage: /part")
				}else{
					// 
					// 1. Find the original channel an delete the user
					oriChannel := server.allUser[u].currentChannel
					delete(server.allChannel[oriChannel].userMap, u)
					// 2. server.allUser[u].currentChannel => default
					// 2-1. Check default channel: if exist add user, if not create one
					if ch, ok := server.allChannel["default"]; ok{
						ch.userMap[u] = server.allUser[u]
					}else{
						// create new channel:
						server.allChannel["default"] = new(Channel)
						server.allChannel["default"].channelName = "default"
						server.allChannel["default"].userMap = make(map[string]*User)
						server.allChannel["default"].userMap[u] = server.allUser[u]
					}
					server.allUser[u].currentChannel = "default"
				}
			/////////// NORMAL MESSAGE:
			default: // set msg to all user in the same channel
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
		allNick:	make(map[string]*User),
		allChannel: make(map[string]*Channel),
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
