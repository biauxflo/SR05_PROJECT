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

// Initialisation des variables
var globalStock int = 300
var localStock int = 50
var seuil int = 25
var siteId int
var pendingRequest bool
var snapshotInProgress bool

// Traitement du stock local et global lors de l'accès à la section critique.
func handleSCStart() {

	retrait := 25
	if globalStock > 25 {
		globalStock = globalStock - retrait
		localStock = localStock + retrait
	} else {
		retrait = globalStock
		globalStock = globalStock - retrait
		localStock = localStock + retrait
	}

	l := log.New(os.Stderr, "", 0)
	l.Println("Site n° " + strconv.Itoa(siteId) + "-> Livraison effectuée : Nouveau stock : " + strconv.Itoa(localStock) + "- Nouveau stock global: " + strconv.Itoa(globalStock))

	//Envoi de la libération de la section critique.
	utils.Send(utils.SCEnd, siteId, siteId, -1, globalStock, utils.Neutre, nil, 0, "")

	pendingRequest = false
}

// Controle de la quantité de stock en cours d'utilisation,
// quand elle descends sous le seuil, une requête est envoyée pour demander une livraison depuis le stock global.
func compare_seuil_stock() {

	for {
		if localStock > 0 {
			localStock--
		}

		r := rand.Intn(5)
		time.Sleep(time.Duration(r) * time.Second)

		if localStock < seuil && !pendingRequest {
			utils.Send(utils.SCRequest, siteId, siteId, -1, -1, utils.Neutre, nil, 0, "")
			pendingRequest = true
		}

		// Lancement de la snapshot en cas d'arrivée à zéro du stock global et local.
		if globalStock == 0 && !snapshotInProgress && siteId == 1 {
			utils.Send(utils.SnapStart, siteId, siteId, localStock, globalStock, utils.Neutre, nil, 0, "")
			snapshotInProgress = true
		}
	}
}

// Traitement de la mise à jour de la section critique par un autre site.
func handleUpdate(newStock int) {
	globalStock = newStock
	l := log.New(os.Stderr, "", 0)
	l.Println("Stock global mis à jour sur le site n°" + strconv.Itoa(siteId) + ":  " + strconv.Itoa(globalStock))
}

// Traitement d'une requête de stock, on utilise le champ d'horloge pour transmettre le stock local car le champ n'est pas utilisé.
func handleStockRequest() {
	utils.Send(utils.StockRequest, siteId, siteId, localStock, globalStock, utils.Neutre, nil, 0, "")
}

// Traitement du message en fonction de son type.
func handleMessage(message utils.Message) {
	switch message.Type {
	case utils.SCStart:
		handleSCStart()
	case utils.SCUpdate:
		handleUpdate(message.GlobalStock)
	case utils.StockRequest:
		handleStockRequest()
	}
}

// Routine pour l'attente de la réception de message.
func waitMessages() {
	for {
		message := utils.Receive()
		if message.Sender == siteId && message.Receiver == siteId {
			handleMessage(message)
		}
	}
}

// Parsing des options
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
