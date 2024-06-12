// main.go

package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"jitsi-auth0-service/platform"
        "jitsi-auth0-service/platform/authenticator"
	"jitsi-auth0-service/platform/router"
)

func main() {
        address := "192.81.134.144:3033"
        certPath := "/etc/ssl/hook/fullchain.pem"
        keyPath := "/etc/ssl/hook/privkey.pem"

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load the env vars: %v", err)
	}

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

        ipStore := platform.NewIPStore()  // Assuming IPStore is part of the authenticator package

	rtr := router.New(auth, ipStore)

	log.Print("Server listening on http://hook.obscurenetworks.com:3033/")
	if err := http.ListenAndServeTLS(address, certPath, keyPath, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
