package main

import (
	"flag"
	"fmt"
	"os"
)

type MessageType int

const (
	Release MessageType = iota
	ACK
	Request
	SCRequest
	SCStart
	SCEnd
)

type Message struct {
	Type       MessageType
	Sender     int
	Receiver   int
	ClockValue int
}

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
		if (k != siteId && Tab[k-1].Type == Request) && (((Tab[k-1].ClockValue) == (Tab[nbSite-1].ClockValue) && (k < nbSite)) || ((Tab[k-1].ClockValue) < (Tab[nbSite-1].ClockValue))) {
			return false
		}
	}

	return true
}

func handleSCRequest() {
	horloge++
	Tab[nbSite-1] = Record{Request, horloge}
	// TODO : SendAll(request,horloge)
}

func handleSCRelease(stock int) {
	horloge++
	Tab[nbSite-1] = Record{Release, horloge}
	// TODO : SendAll(release,horloge,stock)
}

func handleRequest(h int, sender int) {
	horloge = max(horloge, h) + 1
	Tab[sender-1] = Record{Request, h}
	// TODO : Send(ack, siteId, sender, horloge)
	if canEnterCriticalSection() {
		// TODO : Send appli DébutSC
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
		// TODO : Send appli DébutSC
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
		// TODO : Send appli DébutSC
	}
}

func handleMessage(message string) {
	// Décoder le message
	// TODO
	msgType, h, sender, stockValue := decodeMessage(message)

	// Traiter le message en fonction de son type
	switch msgType {
	case SCRequest:
		handleSCRequest()
	case SCEnd:
		handleSCRelease(stockValue)
	case Request:
		handleRequest(h, sender)
	case Release:
		handleRelease(h)
	case ACK:
		handleAck(h, sender)
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

	// Lancer la goroutine pour lire les messages sur stdin
	// TODO : go readInput()

	// Attendre indéfiniment
	select {}
}
