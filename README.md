# Build minimal IRC Server in Go language

### 42 School [Rush project]

## Goal: 

Discover Go and enrich concurrent programming skills by creating IRC server.

> [IRC standard](https://tools.ietf.org/html/rfc1459)
>

## General instructions

1. Only the Go standard library is allowed for this project.

2. You have to handle errors in a sensitive manner.

## Mandatory part

1. Write an IRC server using only the Go language.

2. The server should use concurrency and goroutines to handle multiple clients.

3. Users must be able to sign up on your server through a client with a unique username and password.

4. Users must also have a unique but modifiable nickname. However, usernames can- not change once they are created.

5. User information (nickname, username, and password) should be kept in-memory.

6. Channels are also stored in-memory.

7. If the server shuts down, all channel and user information must be wiped.

8. Your server must support the following commands:

> 8-1. PASS NICK USER - Initial authentication for a user.
>
> 8-2. NICK - Change nickname
>
> 8-3. JOIN - Makes the user join a channel. If the channel doesn’t exist, it will be created.
>
> 8-4. PART - Makes the user leave a channel.
> 
> 8-5. NAMES - Lists all users connected to the server (bonus: make it RFC compliant with channel modes).
> 
> 8-6. LIST - Lists all channels in the server (bonus: make it RFC compliant with channel modes).
>
> 8-7. PRIVMSG - Send a message to another user or a channel.\

[more detail...PDF](https://github.com/AmberFu/frozen_go/blob/master/frozen.en.pdf)

## Usage: go run server.go

1. RUN server: go run server.go & (keep in background)

2. Client: nc localhost 9000 (This server use port 9000)

## Project overview:

1. Create 4 struct: Server, User, Channel, Msg.

2. Use `go handleConnection()` do concurrency.

3. Save each user's connection to communicate with group or with other user.
