package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
	"utils"
)

// Initialiser les variables
var globalStock int = 180
var localStock int = 10
var seuil int = 5
var siteId int
var pendingRequest bool

func handleSCStart() {
	//Traitement du stock local et global

	globalStock = globalStock - 10
	localStock = localStock + 10

	//Envoi de la libération
	utils.Send(utils.SCEnd, siteId, siteId, -1, globalStock)

	pendingRequest = false
}

func compare_seuil_stock() {

	for {
		localStock--
		l := log.New(os.Stderr, "", 0)
		l.Println(strconv.Itoa(siteId) + " " + strconv.Itoa(localStock))
		r := rand.Intn(5)
		time.Sleep(time.Duration(r) * time.Second)
		if localStock < seuil && !pendingRequest {
			utils.Send(utils.SCRequest, siteId, siteId, -1, -1)
			pendingRequest = true
		}
		if localStock <= 0 {
			os.Exit(0)
		}
	}
}

func handleRelease(newStock int) {
	globalStock = newStock
}

func handleMessage(message utils.Message) {

	// Traiter le message en fonction de son type
	switch message.Type {
	case utils.SCStart:
		handleSCStart()
	case utils.Release:
		handleRelease(message.GlobalStock)
	}
}

func waitMessages() {
	for {
		message := utils.Receive()
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
