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
var mat [][]Record
var horloge int
var nbSite int
var siteId int

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

func handleSCRequest() {
	horloge++
	Tab[nbSite-1] = Record{utils.Request, horloge}
	utils.SendAll(utils.ACK, siteId, horloge, -1)
}

func handleSCRelease(stock int) {
	horloge++
	Tab[nbSite-1] = Record{utils.Release, horloge}
	utils.SendAll(utils.Release, siteId, horloge, stock)
}

func handleRequest(h int, sender int) {
	horloge = max(horloge, h) + 1
	Tab[sender-1] = Record{utils.Request, h}
	utils.Send(utils.ACK, siteId, sender, horloge, -1)
	if canEnterCriticalSection() {
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1)
	}
}

// Fonction pour gérer la réception d'un message de libération
func handleRelease(h int) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab
	Tab[siteId-1] = Record{utils.Release, h}

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1)
	}
}

// Fonction pour gérer la réception d'un message d'accusé
func handleAck(h int, sender int) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab ssi il n'est pas en attente de requete
	if Tab[sender-1].Type != utils.Request {
		Tab[sender-1] = Record{utils.ACK, h}
	}

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1)
	}
}

func handleMessage(message utils.Message) {

	// Traiter le message en fonction de son type
	switch message.Type {
	case utils.SCRequest:
		handleSCRequest()
	case utils.SCEnd:
		handleSCRelease(message.GlobalStock)
	case utils.Request:
		handleRequest(message.ClockValue, message.Sender)
	case utils.Release:
		handleRelease(message.Sender)
	case utils.ACK:
		handleAck(message.ClockValue, message.Sender)
	}
}

func waitMessages() {
	for {
		message := utils.Receive()
		if (message.Receiver == 0 && message.Sender != siteId) || message.Receiver == siteId {
			handleMessage(message)
		}
		if mustForward(message) {
			utils.Forward(message)
		}
	}
}

func request() {
	time.Sleep(1000)
	if siteId == 1 {
		utils.SendAll(utils.Request, siteId, 1, 0)
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
	mat = make([][]MaStructure, nbSite)
	for k := 0; k < nbSite; k++ {
		Tab[k] = Record{utils.Release, 0}
		matrice[i] = make([]MaStructure, nbSite)
	}

	// Initialiser l'horloge
	horloge = 0

	go waitMessages()

	request()

	select {}
}
