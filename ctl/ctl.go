package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"utils"
)

var Tab []utils.Record
var horloge int
var nbSite int
var siteId int
var couleur utils.Couleur
var initiateur bool
var bilan int
var EG [][]utils.Record
var stocks []int
var prepost []string
var NbEtatsAttendus int
var NbMessagesAttendus int

func printTab() {
	l := log.New(os.Stderr, "", 0)
	l.Println(strconv.Itoa(siteId) + ": ")
	for k := 0; k < nbSite; k++ {
		l.Print(Tab[k])
	}
}

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
		isSmallestEstampille := ((Tab[k-1].ClockValue) == (Tab[siteId-1].ClockValue) && (k < siteId)) || ((Tab[k-1].ClockValue) < (Tab[siteId-1].ClockValue))
		if k != siteId && (isSmallestEstampille) {
			return false
		}
	}

	return true
}

func mustForward(message utils.Message) bool {
	return message.Sender != siteId && message.Receiver != siteId
}

func handleSCRequest() {
	horloge++
	Tab[siteId-1] = utils.Record{utils.Request, horloge}
	utils.SendAll(utils.Request, siteId, horloge, -1)
	bilan = bilan + nbSite - 1
}

func handleSCRelease(stock int) {
	horloge++
	Tab[siteId-1] = utils.Record{utils.Release, horloge}
	utils.SendAll(utils.Release, siteId, horloge, stock)
}

func handleRequest(h int, sender int) {
	horloge = max(horloge, h) + 1
	Tab[sender-1] = utils.Record{utils.Request, h}
	utils.Send(utils.ACK, siteId, sender, horloge, -1, couleur, Tab, bilan, "")
	bilan++
	if canEnterCriticalSection() {
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, couleur, Tab, bilan, "")
		bilan++
	}
}

// Fonction pour gérer la réception d'un message de libération
func handleRelease(h int, sender int, globalStock int) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab
	Tab[sender-1] = utils.Record{utils.Release, h}

	// Mettre à jour la valeur du stock dans l'application
	utils.Send(utils.SCUpdate, siteId, siteId, horloge, globalStock, couleur, Tab, bilan, "")

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, couleur, Tab, bilan, "")
		bilan++
	}
}

// Fonction pour gérer la réception d'un message d'accusé
func handleAck(h int, sender int) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab ssi il n'est pas en attente de requete
	if Tab[sender-1].Type != utils.Request {
		Tab[sender-1] = utils.Record{utils.ACK, h}
	}

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, couleur, Tab, bilan, "")
		bilan++
	}
}

func startSnapShot() {
	couleur = utils.Rouge
	initiateur = true
	EG[siteId-1] = Tab
	NbEtatsAttendus = nbSite - 1
	NbMessagesAttendus = bilan
	// TODO : utils.Send(utils.StockRequest, siteId, siteId, -1, -1)
	bilan++
}

func handleEtat(sender int, etat []utils.Record, bilan int) {
	NbMessagesAttendus += bilan
	EG[sender-1] = etat
	NbEtatsAttendus--

	if NbMessagesAttendus == 0 && NbEtatsAttendus == 0 {
		// TODO : FIN
	}
}

func handlePrepost(sender int, prepostMessage string) {
	NbMessagesAttendus--
	prepost[sender] += prepostMessage + "\n"
}

func handleMessage(message utils.Message) {
	// Traiter le message en fonction de son type
	switch message.Type {
	case utils.SCRequest:
		handleSCRequest()
	case utils.SCEnd:
		bilan--
		handleSCRelease(message.GlobalStock)
	case utils.Request:
		bilan--
		handleRequest(message.ClockValue, message.Sender)
	case utils.Release:
		handleRelease(message.ClockValue, message.Sender, message.GlobalStock)
	case utils.ACK:
		bilan--
		handleAck(message.ClockValue, message.Sender)
	case utils.SnapStart:
		startSnapShot()
	case utils.Etat:
		handleEtat(message.Sender, message.Etat, message.Bilan)
	case utils.Prepost:
		handlePrepost(message.Sender, message.PrepostMessage)
	}
}

func waitMessages() {
	for {
		message := utils.Receive()
		l := log.New(os.Stderr, "", 0)
		l.Println(strconv.Itoa(siteId) + " <-- Type : " + strconv.Itoa(int(message.Type)) + " Sender :  " + strconv.Itoa(message.Sender) + " Clock : " + strconv.Itoa(message.ClockValue))

		if message.Couleur == utils.Blanc && couleur == utils.Rouge {
			utils.Send(utils.SCStart, siteId, siteId, horloge, -1, couleur, Tab, bilan, utils.EncodeSimpleMessage(message))
		}

		if message.Couleur == utils.Rouge && couleur == utils.Blanc {
			couleur = utils.Rouge
			utils.Send(utils.Etat, siteId, siteId, horloge, -1, couleur, Tab, bilan, "")
		}

		if initiateur == true {
			switch message.Type {
			case utils.Prepost:
				handlePrepost(message.Sender, message.PrepostMessage)
			case utils.Etat:
				handleEtat(message.Sender, message.Etat, message.Bilan)
			}
		} else {
			if (message.Receiver == 0 && message.Sender != siteId) || message.Receiver == siteId {
				handleMessage(message)
			}
		}
		if mustForward(message) {
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
	Tab = make([]utils.Record, nbSite)
	for k := 0; k < nbSite; k++ {
		Tab[k] = utils.Record{utils.Release, 0}
	}

	// Initialiser l'horloge
	horloge = 0

	go waitMessages()

	select {}
}
