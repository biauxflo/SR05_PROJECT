package app

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// Initialiser les variables
var globalStock int = 180
var localStock int = 40
var seuil int = 25
var siteId int
var pendingRequest bool

func handleSCStart() {
	//Traitement du stock local et global

	globalStock = globalStock - 10
	localStock = localStock + 10

	//Envoi de la libération
	Send(SCEnd, siteId, siteId, -1, globalStock)

	pendingRequest = false
}

func compare_seuil_stock() {

	for {
		localStock--
		time.Sleep(time.Duration(rand.Int() * 100))
		if localStock < seuil && !pendingRequest {
			Send(SCRequest, siteId, siteId, -1, -1)
			pendingRequest = true
		}
	}
}

func handleRelease(newStock int) {
	globalStock = newStock
}

func handleMessage(msgType MessageType) {

	// Traiter le message en fonction de son type
	switch msgType {
	case SCStart:
		handleSCStart()
	case Release:
		handleRelease(newStock)
	}
}

func waitMessages() {
	for {
		message := Receive()
		if message.Sender == siteId && message.Receiver == siteId {
			handleMessage(message)
		}
	}
}

func main() {
	flag.IntVar(&siteId, "n", 0, "Numéro du site à contrôler.")
	flag.Parse()

	if siteId == 0 {
		fmt.Println("Erreur lors de la sélection du site à controler (-n)")
		os.Exit(1)
	}

	pendingRequest = false

	// Lancement des deux go routines

	go waitMessages()

	go compare_seuil_stock()

	// Attendre indéfiniment
	select {}
}
