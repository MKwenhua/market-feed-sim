package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"
	"symbolsock"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
)
const upgrader = websocket.Upgrader{}



func main() {
	http.HandleFunc("/", chillin)
	
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
	panic(err)
	}
	defer c.Close()
		
	theData, err := redis.String(c.Do("GET", "eur_usd_m1"))
	if err != nil {
	  println("key not found")
	}

	seriesSockets := symbolsock.SymbolStream(theData)
	seriesSockets.newStream("APPL")
	for i:= 0; i < 6; i++{
		println(seriesSockets.getSeriesPoint("APPL"))
	}
    
		


	 http.HandleFunc("/indi", func(w http.ResponseWriter, r *http.Request){
		http.ServeFile(w, r, "index.html")
     })
	 http.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request){
		var conn, _ = upgrader.Upgrade(w, r, nil)
		
		go func(conn *websocket.Conn) { 
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					conn.Close()
				}
			}

		}(conn)
		go func(conn *websocket.Conn) {
			ch := time.Tick(1 *time.Second)
			for range ch {
				
			 	conn.WriteJSON({yo: "hey"})
			}
		}(conn)
    })
	http.ListenAndServe(":5050",nil)
	
		
}

func chillin(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "mirco service, chillin. Running version %s", runtime.Version())
}