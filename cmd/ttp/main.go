/// main.go
/// Punkt wejścia aplikacji Trusted Third Party.
///
/// Plik uruchamia usługę TTP odpowiedzialną za rejestrację aplikacji,
/// wydawanie certyfikatów X.509 oraz udział w uwierzytelnianiu klienta i serwera.

package main

import (
	"log"
	thirdpart "scs/internal/third-part"
)

/// @brief Uruchamia aplikację TTP.
///
/// Funkcja tworzy instancję TTP na podstawie zmiennych środowiskowych
/// i uruchamia usługę odpowiedzialną za obsługę protokołu zaufanej strony trzeciej.

func main() {
	app, err := thirdpart.NewAppFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	if err = app.Run(); err != nil {
		log.Fatal(err)
	}
}
