package ctl

import (
	"flag"
	"fmt"
	"os"
	"time"
	"utils"
)

type Record struct {
	Type       utils.MessageType
	ClockValue int
}

var Tab []Record
var EG [][]Record
var GlobalStocks []int
var PrePost []utils.MessageType
var horloge int
var nbSite int
var siteId int
var couleur utils.Couleur
var bilan int
var nbEtatsAttendus int
var nbMsgAttendus int
var snapshotIsFinished bool


func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func canEnterCriticalSection() bool {

	if Tab[siteId-1].Type != utils.Request {
		return false
	}

	// Check if there is an older request pending
	for k := 1; k <= nbSite; k++ {
		isACKorRequest := Tab[k-1].Type == utils.Request || Tab[k-1].Type == utils.ACK
		isSmallestEstampille := ((Tab[k-1].ClockValue) == (Tab[nbSite-1].ClockValue) && (k < nbSite)) || ((Tab[k-1].ClockValue) < (Tab[nbSite-1].ClockValue))
		if (k != siteId && isACKorRequest) && (isSmallestEstampille) {
			return false
		}
	}

	return true
}

func mustForward(message utils.Message) bool {
	return message.Sender != siteId
}

func handleSCRequest(color utils.Couleur) {
	horloge++
	Tab[siteId-1] = Record{utils.Request, horloge}
	utils.SendAll(utils.Request, siteId, horloge, -1, color, utils.Request)
}

func handleSCRelease(stock int, color utils.Couleur) {
	horloge++
	Tab[siteId-1] = Record{utils.Release, horloge}
	utils.SendAll(utils.Release, siteId, horloge, stock, color, utils.Release)
}

func handleRequest(h int, sender int, color utils.Couleur) {
	horloge = max(horloge, h) + 1
	Tab[sender-1] = Record{utils.Request, h}
	utils.Send(utils.ACK, siteId, sender, horloge, -1, color, utils.ACK)
	if canEnterCriticalSection() {
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, color, utils.SCStart)
	}
}

// Fonction pour gérer la réception d'un message de libération
func handleRelease(h int, color utils.Couleur) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab
	Tab[siteId-1] = Record{utils.Release, h}

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, color, utils.SCStart)
	}
}

// Fonction pour gérer la réception d'un message d'accusé
func handleAck(h int, sender int, color utils.Couleur) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab ssi il n'est pas en attente de requete
	if Tab[sender-1].Type != utils.Request {
		Tab[sender-1] = Record{utils.ACK, h}
	}

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, color, utils.SCStart)
	}
}

//fonction permettant d'alterner entre les couleur d'un site
func changeColor() {
    if couleur == utils.Blanc {
        couleur = utils.Rouge
    } else {
        couleur = utils.Blanc
    }
}

// Fonction pour gérer la réception d'un message de début de snapshot
func handleSnapStart(stock int) {
	//si une snapshot n'est pas en cours et que l'on est bien le site 1 qui s'occupe des snapshots
	if (snapshotIsFinished && siteId == 1){
		snapshotIsFinished = false
		//on change de couleur pour signaliser une nouvelle snapshot
		changeColor()

		//on stock l'état local
		EG[siteId-1] = Tab
		GlobalStocks[siteId-1] = stock

		nbEtatsAttendus = nbSite - 1
		nbMsgAttendus = bilan

	}
}

func handleMessage(message utils.Message) {

	// Traiter le message en fonction de son type
	switch message.Type {
	case utils.SCRequest:
		handleSCRequest(message.Color)
	case utils.SCEnd:
		handleSCRelease(message.GlobalStock,message.Color)
	case utils.Request:
		handleRequest(message.ClockValue, message.Sender,message.Color)
	case utils.Release:
		handleRelease(message.Sender,message.Color)
	case utils.ACK:
		handleAck(message.ClockValue, message.Sender,message.Color)
	case utils.Etat:
		handleEtat()
	case utils.PrePost:
		handlePrePost()
	case utils.SnapStart:
		handleSnapStart(message.GlobalStock)
	}
}

func waitMessages() {
	for {
		message := utils.Receive()
		if (message.Receiver == 0 && message.Sender != siteId) || message.Receiver == siteId {
			if message.Type != utils.SnapStart || message.Type != utils.PrePost || message.Type != utils.Etat{
					bilan--
			}
			handleMessage(message)
		}
		if mustForward(message) {
			if message.Type != utils.PrePost || message.Type != utils.Etat{
					bilan++
			}
			utils.Forward(message)
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

	// Initialiser le programme
	nbSite = 3 // number of sites
	Tab = make([]Record, nbSite)
	EG = make([][]Record, nbSite)
	GlobalStocks = make([]int, nbSite)
	for k := 0; k < nbSite; k++ {
		Tab[k] = Record{utils.Release, 0}
		EG[k] = make([]Record, nbSite)
		GlobalStocks[k] = 0
	}

	// Initialiser l'horloge
	horloge = 0
	//initialiser la couleur du site
	couleur = utils.Blanc
	//initialisations pour la snapshot
	bilan = 0
	snapshotIsFinished = true
	nbMsgAttendus = 0
	nbEtatsAttendus = 0

	go waitMessages()


	select {}
}
