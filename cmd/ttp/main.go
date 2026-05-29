package main

import (
	"log"
	thirdpart "scs/internal/third-part"
)

func main() {
	app, err := thirdpart.NewAppFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	if err = app.Run(); err != nil {
		log.Fatal(err)
	}
}
