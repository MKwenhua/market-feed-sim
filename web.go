package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"
	"ticksimulator"
	"github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{}


 
func main() {
	http.HandleFunc("/", hello)
	
	onOpenShift := false
    if ( onOpenShift) {
		bind := fmt.Sprintf("%s:%s", os.Getenv("OPENSHIFT_GO_IP"), os.Getenv("OPENSHIFT_GO_PORT"))
		fmt.Printf("listening on %s...", bind)
		err := http.ListenAndServe(bind, nil)
		if err != nil {
			panic(err)
		}
	}
	
		


	 http.HandleFunc("/indi", func(w http.ResponseWriter, r *http.Request){
		http.ServeFile(w, r, "index.html")
     })
	 http.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request){
		var conn, _ = upgrader.Upgrade(w, r, nil)
		
		go func(conn *websocket.Conn) { 
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					conn.Close()
				}
				str :=  string(msg[:])
				println(string(str))
				symbolStream(conn, str)
				
			}

		}(conn)
	
    })
	http.ListenAndServe(":5050",nil)
	
		
}
func symbolStream(conn *websocket.Conn, symb string){
	seriesPush := ticksimulator.GetTickStruct(symb)
	
	ch := time.Tick(1 *time.Second)
		for range ch {
			update := seriesPush.GetPoint()
		 	conn.WriteJSON(update)
	}
	
}
func hello(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(res, "hello, world from %s", runtime.Version())
}