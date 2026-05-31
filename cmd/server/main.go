/// main.go
/// Punkt wejścia aplikacji serwera.
///
/// Plik uruchamia aplikację serwera, która udostępnia usługę klientowi
/// oraz współpracuje z TTP podczas procesu uwierzytelniania.

package main

import (
	"log"
	"scs/internal/server"
)

// / @brief Uruchamia aplikację serwera.
// /
// / Funkcja tworzy instancję serwera na podstawie zmiennych środowiskowych,
// / wykonuje inicjalizację oraz uruchamia główną aplikację serwerową.
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
