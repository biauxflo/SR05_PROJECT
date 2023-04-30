package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

type Record struct {
	Type       MessageType
	ClockValue int
}

var Tab []Record
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
	// Check if there is an older request pending
	for k := 1; k <= nbSite; k++ {
		isACKorRequest := Tab[k-1].Type == Request || Tab[k-1].Type == ACK
		isSmallestEstampille := ((Tab[k-1].ClockValue) == (Tab[nbSite-1].ClockValue) && (k < nbSite)) || ((Tab[k-1].ClockValue) < (Tab[nbSite-1].ClockValue))
		if (k != siteId && isACKorRequest) && (isSmallestEstampille) {
			return false
		}
	}

	return true
}

func mustForward(message Message) bool {
	return message.Sender != siteId
}

func handleSCRequest() {
	horloge++
	Tab[nbSite-1] = Record{Request, horloge}
	SendAll(ACK, siteId, horloge, -1)
}

func handleSCRelease(stock int) {
	horloge++
	Tab[nbSite-1] = Record{Release, horloge}
	SendAll(Release, siteId, horloge, stock)
}

func handleRequest(h int, sender int) {
	horloge = max(horloge, h) + 1
	Tab[sender-1] = Record{Request, h}
	Send(ACK, siteId, sender, horloge, -1)
	if canEnterCriticalSection() {
		Send(SCStart, siteId, siteId, horloge, -1)
	}
}

// Fonction pour gérer la réception d'un message de libération
func handleRelease(h int) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab
	Tab[siteId-1] = Record{Release, h}

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		Send(SCStart, siteId, siteId, horloge, -1)
	}
}

// Fonction pour gérer la réception d'un message d'accusé
func handleAck(h int, sender int) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab ssi il n'est pas en attente de requete
	if Tab[sender-1].Type != Request {
		Tab[sender-1] = Record{ACK, h}
	}

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		Send(SCStart, siteId, siteId, horloge, -1)
	}
}

func handleMessage(message Message) {

	// Traiter le message en fonction de son type
	switch message.Type {
	case SCRequest:
		handleSCRequest()
	case SCEnd:
		handleSCRelease(message.GlobalStock)
	case Request:
		handleRequest(message.ClockValue, message.Sender)
	case Release:
		handleRelease(message.Sender)
	case ACK:
		handleAck(message.ClockValue, message.Sender)
	}
}

func waitMessages() {
	for {
		message := Receive()
		if message.Receiver == 0 || message.Receiver == siteId {
			handleMessage(message)
		}
		if mustForward(message) {
			Forward(message)
		}
	}
}

func request() {
	time.Sleep(1000)
	if siteId == 1 {
		SendAll(Request, siteId, 1, 0)
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
	for k := 0; k < nbSite; k++ {
		Tab[k] = Record{Release, 0}
	}

	// Initialiser l'horloge
	horloge = 0

	go waitMessages()

	request()
}
