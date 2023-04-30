package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"./MsgUtil"
)

type Record struct {
	Type       MsgUtil.MessageType
	ClockValue int
}

var Tab []Record
var horloge int
var globalStock int
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
		if (k != siteId && Tab[k-1].Type == MsgUtil.Request) && (((Tab[k-1].ClockValue) == (Tab[nbSite-1].ClockValue) && (k < nbSite)) || ((Tab[k-1].ClockValue) < (Tab[nbSite-1].ClockValue))) {
			return false
		}
	}

	return true
}

func handleSCRequest() {
	horloge++
	Tab[nbSite-1] = Record{MsgUtil.Request, horloge}
	MsgUtil.SendAll(siteId, horloge)
}

func handleSCRelease(sender int, receiver int, stock int) {
	horloge++
	Tab[nbSite-1] = Record{MsgUtil.Release, horloge}
	globalStock -= stock
	if siteId == receiver {
		MsgUtil.SendAllRelease(sender, receiver, stock)
	} else if siteId != sender {
		MsgUtil.Send(sender, horloge, receiver, stock)
	}
}

func handleRequest(h int, sender int) {
	horloge = max(horloge, h) + 1
	Tab[sender-1] = Record{MsgUtil.Request, h}
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
	Tab[siteId-1] = Record{MsgUtil.Release, h}

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
	if Tab[sender-1].Type != MsgUtil.Request {
		Tab[sender-1] = Record{MsgUtil.ACK, h}
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
	msgType, _ := MsgUtil.DecodeMessageType(MsgUtil.Findval(message, "Type"))
	h, _ := strconv.Atoi(MsgUtil.Findval(message, "ClockValue"))
	sender, _ := strconv.Atoi(MsgUtil.Findval(message, "Sender"))
	stockValue, _ := strconv.Atoi(MsgUtil.Findval(message, "StockValue"))

	// Traiter le message en fonction de son type
	switch msgType {
	case MsgUtil.SCRequest:
		handleSCRequest()
	case MsgUtil.SCEnd:
		handleSCRelease(siteId, h, stockValue)
	case MsgUtil.Request:
		handleRequest(h, sender)
	case MsgUtil.Release:
		handleRelease(h)
	case MsgUtil.ACK:
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
		Tab[k] = Record{MsgUtil.Release, 0}
	}

	// Initialiser l'horloge
	horloge = 0

	// Initialiser le stock
	globalStock = 50

	// Lancer la goroutine pour lire les messages sur stdin
	go MsgUtil.Receive()

	// Attendre indéfiniment
	select {}
}
