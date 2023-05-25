package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Msg struct {
	id  string
	msg string
}

type Client chan Msg

var (
	clients = make(map[Client]bool)
)

func main() {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	r.Get("/sse", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported!", http.StatusInternalServerError)
			return
		}

		id := r.URL.Query().Get("id")

		client := make(Client)
		clients[client] = true

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Connection", "keep-alive")

		for {
			select {
			case <-r.Context().Done():
				fmt.Println("a client stopped listening", id)
				delete(clients, client)
				return
			case data := <-client:
				if data.id == id {
					fmt.Fprintf(w, "event: custom\ndata: %v\n\n", data.msg)
					flusher.Flush()
				}
			}
		}
	})

	go func() {
		for {
			id := strconv.Itoa(rand.Intn(3))
			for client := range clients {
				client <- Msg{
					id:  id,
					msg: "hello " + id + " " + time.Now().Format(time.RFC3339),
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Serving http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
