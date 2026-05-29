package main

import (
	"log"
	"scs/internal/server"
)

func main() {
	app, err := server.NewAppFromEnv()
	if err != nil {
		log.Fatalln(err)
	}
	if err = app.Bootstrap(); err != nil {
		log.Fatalln(err)
	}
	if err = app.Run(); err != nil {
		log.Fatalln(err)
	}
}
