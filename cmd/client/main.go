package main

import (
	"log"

	"scs/internal/client"
)

func main() {
	app, err := client.NewAppFromEnv()
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
