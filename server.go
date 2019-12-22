package main

import (
	"fmt"
	"net"
	"io"
	"bufio"
	"log"
)

////////////////////////////////////
// const:
////////////////////////////////////
const max_user int = 100

////////////////////////////////////
// struct:
////////////////////////////////////
type User struct {
	uName string
	uNick string
	passward string
	currentChannel string
	message chan string
	//connect conn
}

type Channel struct {
	channelName string
	userArray []string
	message chan string
}

// Create map of unique user name:
var userMap = map[string]*User{}

// Create map of unique nick name: (only for check)
var nickMap = map[string]bool{}

// Create map of unique channel name as key and string array as array:
var channelMap = map[string]*Channel{}

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

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// 1. INPUT user name:
	io.WriteString(conn, "Please enter your name: ")
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	uname := only_take_isprint(scanner.Text())
	// if user name already exist:
	// 		- check password
	//		- sent user into chat room
	if existUser, ok := userMap[uname]; ok {
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
		userMap[uname] = new(User)
		userMap[uname].uName = uname
		// 2. INPUT nick name:
		io.WriteString(conn, "Please enter your nick name: ")
		scanner.Scan()
		nickname := scanner.Text()
		// -> use nickMap to check if the nickName has been taken:
		_, nickOk := nickMap[nickname]
		for nickOk {
			io.WriteString(conn, "This nickName has been taken, please choose another nick name: ")
			scanner.Scan()
			nickname := scanner.Text()
			_, nickOk = nickMap[nickname]
		}
		nickMap[nickname] = true
		userMap[uname].uNick = nickname
		// 3. INPUT password:
		io.WriteString(conn, "Please enter your password: ")
		scanner.Scan()
		inputPassword := scanner.Text()
		userMap[uname].passward = inputPassword
		fmt.Println("userMap = \n", userMap) // for debug...
		fmt.Println("nickMap = \n", nickMap) // for debug...
	}
	// 4. ENTER chat room:
	io.WriteString(conn, "Please enter chat room: ")
	scanner.Scan()
	chatroom := scanner.Text()
	fmt.Println(chatroom)
	// If the chat room is exist:
	if chatRoom, ok := channelMap[chatroom]; ok {
		// update user info of currentChannel
		userMap[uname].currentChannel = chatroom
		// add user into channel (Q: do I need to check user first?!?!?!)
		chatRoom.userArray = append(chatRoom.userArray, uname)
	// If the chat room is not exist:
	} else {
		channelMap[chatroom] = new(Channel)
		channelMap[chatroom].channelName = chatroom
		channelMap[chatroom].userArray = append(channelMap[chatroom].userArray, uname)
	}
	io.WriteString(conn, "\n" + uname + "join the " + chatroom + "\n")

	// Wait user to input: -> maybe create a real function... and use gorutine
	func () {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}
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
		log.Println("err")
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			log.Println("err")
		}
		go handleConnection(conn)
	}
}
