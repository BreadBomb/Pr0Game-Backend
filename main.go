package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/googollee/go-socket.io"
)

type customServer struct {
	Server *socketio.Server
}

//Header handling, this is necessary to adjust security and/or header settings in general
//Please keep in mind to adjust that later on in a productive environment!
//Access-Control-Allow-Origin will be set to whoever will call the server
func (s *customServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}

func main() {
	//get/configure socket.io websocket for clients
	ioServer := configureSocketIO()

	wsServer := new(customServer)
	wsServer.Server = ioServer

	//HTTP settings
	println("Core Service is listening on port 5000...")
	http.Handle("/socket.io/", wsServer)
	http.ListenAndServe(":5000", nil)
}

func configureSocketIO() *socketio.Server {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	//Client connects to server
	server.On("connection", func(so socketio.Socket) {

		//What will happen as soon as the connection is established:
		so.On("connection", func(msg string) {
			println(so.Id() + " joined clients.")

			//In case you want to send a custom emit directly after the client connected.
			//If you fire an emit directly after the connection event it won't work therefore you need to wait a bit
			//In this case two seconds.
			ticker := time.NewTicker(time.Second)
			go func() {
				for {
					select {
					case <-ticker.C:
						so.Emit("online", "Do Something!")
						err := so.Join("global")
						if err != nil {
							fmt.Println(err)
						}
						ticker.Stop()
						return
					}
				}
			}()
		})

		//What will happen if clients disconnect
		so.On("disconnection", func() {
			log.Println("on disconnect")
		})

		so.On("chat message", func(msg string) {
			fmt.Println("message received... " + msg)
			server.BroadcastTo("global", "message", msg)
		})

		//Custom event as example
		so.On("hello", func(msg string) {
			log.Println("received request (hello): " + msg)

			so.Emit("Hi", "How can I help you?")
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	return server
}