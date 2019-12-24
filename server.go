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

////////////////////////////////////
// functions:
////////////////////////////////////

// not import import...only use in User's uName:
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

// Print out prompt:
func chatFormat(u User) string {
	return "\n["  + u.currentChannel + "] " + u.uNick + " > "
}

// For command /privmsg:
func private_msg(u *User, msg Msg) {
	io.WriteString(u.conn, "\n" + msg.uName + " sent you message: " + msg.says + "\n")
	io.WriteString(u.conn, chatFormat(*u))
}

// Sending message to a group of people:
func public_msg(u User, msg Msg) {
	if msg.uName == ""{
		io.WriteString(u.conn, "\n" + msg.says + "\n")
	}else{
		io.WriteString(u.conn, "\n" + msg.uName + ": " + msg.says + "\n")
	}
	io.WriteString(u.conn, chatFormat(u))
}

// Get all users in current channel without the user who sending message.
// Sub-function for sentChannelMsg():
func getAllUsersConn(targetChannelName string, server Server, self string) []*User {
	var connArray []*User
	// Use channel name to get all user, and append connection into an array:
	for uname, uptr := range server.allChannel[targetChannelName].userMap{
		if uname != self {
			connArray = append(connArray, uptr)
		}
	}
	return connArray
}

// Sub-function for sentChannelMsg():
func sentMultiMsg(channelConnArray []*User, msg Msg){
	for _ , con := range channelConnArray{
		public_msg(*con, msg)
	}
}

// Sending message to specific channel:
func sentChannelMsg(targetChannelName string, server Server, msg Msg, self string){
	connArr := getAllUsersConn(targetChannelName, server, self)
	sentMultiMsg(connArr, msg)
}

// Actually, I don't use this function...
func countUser(server Server) int {
	i := 0
	for range server.allUser {
		i++
	}
	return i
}

// Actually, I don't use this function...
func countChannel(server Server) int {
	i := 0
	for range server.allChannel {
		i++
	}
	return i
}

// MAIN FUNCTION in this project: (concurrent function, the only goroutine I used)
func handleConnection(conn net.Conn, server Server) {
	defer conn.Close()

	fmt.Println("")
	io.WriteString(conn, "\n##### WELCOME TO IRC SERVER #####\n\n")
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
		for inputPassword != existUser.passward {
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
		fmt.Println("From ", conn.RemoteAddr().String(), "\t", uname, " JOIN to OUR IRC SITE!")
		// 2. INPUT nick name:
		io.WriteString(conn, "Please enter your nick name: ")
		scanner.Scan()
		nickname := scanner.Text()
		// -> use nickMap to check if the nickName has been taken:
		_, nickOk := server.allNick[nickname]
		for nickOk {
			io.WriteString(conn, "This nickName has been taken, please choose another nick name: ")
			scanner.Scan()
			nickname = scanner.Text()
			_, nickOk = server.allNick[nickname]
		}
		server.allNick[nickname] = server.allUser[uname]
		server.allUser[uname].uNick = nickname
		// 3. INPUT password:
		io.WriteString(conn, "Please enter your password: ")
		scanner.Scan()
		inputPassword := scanner.Text()
		server.allUser[uname].passward = inputPassword
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
		// sent MSG to the chat room:
		thisMsg := Msg{
			uName: "",
			says: "\n" + uname + " join the channel...\n",
		}
		sentChannelMsg(chatroom, server, thisMsg, uname)
	// If the chat room is not exist:
	} else {
		server.allChannel[chatroom] = new(Channel)
		server.allChannel[chatroom].channelName = chatroom
		server.allChannel[chatroom].userMap = make(map[string]*User)
		server.allChannel[chatroom].userMap[uname] = server.allUser[uname]
		// update user info of currentChannel
		server.allUser[uname].currentChannel = chatroom
		fmt.Println("NEW CHANNEL HAS BEEN CREATED: " + chatroom)
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
					io.WriteString(conn, "Usage: /nick newNickName\n")
				}else{
					oriNick := server.allUser[u].uNick
					newNick := strings.Trim(commandSplit[1], " ")
					_, nickOk := server.allNick[newNick]
					if nickOk {
						io.WriteString(conn, "This nickName has been taken, please try again.")
					}else{
						// 1. modify server.allUser.uNick
						server.allUser[u].uNick = newNick
						// 2. delete server.allNick: delete(m, "route")
						delete(server.allNick, oriNick)
						// 3. add new nickName: server.allNick
						server.allNick[newNick] = server.allUser[uname]
					}
				}
			/////////// JOIN:
			case "/join":
				if len(commandSplit) != 2 {
					io.WriteString(conn, "Usage: /join anotherChannel\n")
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
						fmt.Println("NEW CHANNEL HAS BEEN CREATED: " + chatroom)
					}
					// 3. server.allUser[u].currentChannel => new channel
					server.allUser[u].currentChannel = newChannel
					// 4. Sent Channel message: says good-bye to original channel!
					byeMsg := Msg{
						uName: "",
						says: "\n" + u + " left the channel...\n",
					}
					sentChannelMsg(oriChannel, server, byeMsg, u)
					// 5. Sent Channel message: says Hi to joined channel!
					hiMsg := Msg{
						uName: "",
						says: "\n" + u + " join the channel...\n",
					}
					sentChannelMsg(newChannel, server, hiMsg, u)
				}
			/////////// NAMES: List all nick name
			case "/names": // allChannel map[string]*Channel -> userMap map[string]*User -> uNick
				if len(commandSplit) == 2 {
					f := false
					// finding specific channel and list all user:
					for room, chStruct := range server.allChannel {
						if room == commandSplit[1] {
							f = true
							io.WriteString(conn, "Channel - " + room + "\n")
							for _, Uptr := range chStruct.userMap {
								io.WriteString(conn, "\tUsers: " + Uptr.uNick + "\n")
							}
						}
					}
					if f == false {
						io.WriteString(conn, "Cannot find this channel.\n")
					}
				}else if len(commandSplit) == 1 {
					// List all user's NICK name: seprate by channel
					for room, chStruct := range server.allChannel {
						io.WriteString(conn, "Channel - " + room + "\n")
						for _, Uptr := range chStruct.userMap {
							io.WriteString(conn, "\tUsers: " + Uptr.uNick + "\n")
						}
						io.WriteString(conn, "\n")
					}
				}else{
					io.WriteString(conn, "Usage: /names [channel name]\n")
				}
			/////////// LIST:
			case "/list": // allChannel map[string]*Channel
				if len(commandSplit) != 1 {
					io.WriteString(conn, "Usage: /list\n")
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
					io.WriteString(conn, "Usage: /privmsg NickName Message\n")
				}else{
					// Get this user's nick name:
					myNick := server.allUser[u].uNick
					// Get dst conn:
					sentToUser :=  privArgv[1]
					if usrPtr, ok := server.allNick[sentToUser]; ok{
						// dstConn := usrPtr.conn
						text := privArgv[2]
						thisMsg := Msg{
							uName: myNick,
							says: text,
						}
						private_msg(usrPtr, thisMsg)
					}else{
						io.WriteString(conn, "Sorry! I can't find this user!\nPlease try again!\n")
					}
				}
			/////////// PART:
			case "/part":
				if len(commandSplit) != 1 {
					io.WriteString(conn, "Usage: /part\n")
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
						fmt.Println("NEW CHANNEL HAS BEEN CREATED: default")
					}
					server.allUser[u].currentChannel = "default"
					// 3. Sent Channel message: says user leave the channel!
					thisMsg := Msg{
						uName: "",
						says: "\n" + u + " left the channel...\n",
					}
					sentChannelMsg(oriChannel, server, thisMsg, u)
				}
			/////////// HELP:
			case "/help":
				io.WriteString(conn, "1. use LIST to list all channels\n\tUsage: /list\n")
				io.WriteString(conn, "2. use PART to leave current channel\n\tUsage: /part\n")
				io.WriteString(conn, "3. use NICK to change user's nick name\n\tUsage: /nick newNickName\n")
				io.WriteString(conn, "4. use JOIN to join to another channel\n\tUsage: /join anotherChannel\n")
				io.WriteString(conn, "5. use NAMES to list all user, also can use NAMES CHANNELNAME to list specific channel users\n\tUsage: /names [channel name]\n")
				io.WriteString(conn, "6. use PRINMSG to sent direct message\n\tUsage: /privmsg NickName Message\n")
			/////////// NORMAL MESSAGE:
			default: // set msg to all user in the same channel
				thisMsg := Msg{
					uName: server.allUser[u].uNick,
					says: scanner.Text(),
				}
				sentChannelMsg(server.allUser[u].currentChannel, server, thisMsg, u)
			}
			io.WriteString(conn, chatFormat(*server.allUser[u]))
		}
	}(uname)
}

////////////////////////////////////
// MAIN:
//	1. Listen port 9000
//	2. Create Server struct
//	3. go handleConnection(conn, serverStruct)
//
// 	ps. check the listening port: lsof -nP +c 15 | grep LISTEN
////////////////////////////////////
func main() {
	ln, err := net.Listen("tcp", ":9000")
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
