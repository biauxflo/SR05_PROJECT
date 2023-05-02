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
var globalStock int = 10
var localStock int = 10
var seuil int = 5
var siteId int
var pendingRequest bool
var snapshotInProgress bool

func handleSCStart() {
	//Traitement du stock local et global

	globalStock = globalStock - 10
	localStock = localStock + 10

	l := log.New(os.Stderr, "", 0)
	l.Println(strconv.Itoa(siteId) + "-> Livraison effectuée : Nouveau stock : " + strconv.Itoa(localStock) + "- Nouveau stock global: " + strconv.Itoa(globalStock))

	//Envoi de la libération
	utils.Send(utils.SCEnd, siteId, siteId, -1, globalStock, utils.Neutre, nil, 0, "")

	pendingRequest = false
}

func compare_seuil_stock() {

	for {
		if localStock > 0 {
			localStock--
		}
		l := log.New(os.Stderr, "", 0)
		l.Println(strconv.Itoa(siteId) + " " + strconv.Itoa(localStock))
		r := rand.Intn(5)
		time.Sleep(time.Duration(r) * time.Second)
		if localStock < seuil && !pendingRequest {
			utils.Send(utils.SCRequest, siteId, siteId, -1, -1, utils.Neutre, nil, 0, "")
			pendingRequest = true
		}

		if globalStock == 0 && !snapshotInProgress && siteId == 1 {
			utils.Send(utils.SnapStart, siteId, siteId, -1, -1, utils.Neutre, nil, 0, "")
			snapshotInProgress = true
		}
	}
}

func handleUpdate(newStock int) {
	globalStock = newStock
	l := log.New(os.Stderr, "", 0)
	l.Println("Maj du stock global :  " + strconv.Itoa(globalStock))
}

func handleMessage(message utils.Message) {

	// Traiter le message en fonction de son type
	switch message.Type {
	case utils.SCStart:
		handleSCStart()
	case utils.SCUpdate:
		handleUpdate(message.GlobalStock)
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
