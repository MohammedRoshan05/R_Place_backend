package main

import (
	"fmt"
	"log"

	"github.com/MohammedRoshan05/R_Place_backend/cmd"
)

func main(){
	store,err := cmd.ConnectDB()
	if err != nil {
		fmt.Println("Wow")
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}
	server := &cmd.APIServer{
		ListenAddr: "8000",
		Store: store,
	}	
	server.Run()
}