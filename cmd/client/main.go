/// main.go
/// Punkt wejścia aplikacji klienta.
///
/// Plik uruchamia aplikację klienta, która odpowiada za komunikację
/// z serwerem oraz udział w procesie uwierzytelniania z wykorzystaniem TTP.

package main

import (
	"log"

	"scs/internal/client"
)

// / @brief Uruchamia aplikację klienta.
// /
// / Funkcja tworzy instancję aplikacji na podstawie zmiennych środowiskowych,
// / wykonuje inicjalizację oraz uruchamia główny serwer HTTP klienta.
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
