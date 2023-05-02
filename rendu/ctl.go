package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"utils"
)

// Déclaration des variables.
var Tab []utils.Record
var horloge int
var nbSite int
var siteId int
var couleur utils.Couleur
var initiateur bool
var bilan int
var EG [][]utils.Record
var globalStocks []int
var localStocks []int
var prepost []string
var NbEtatsAttendus int
var NbMessagesAttendus int

// Donne le maximum entre 2 entiers.
func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// Renvoi l'état global du site sous forme de string.
func printEG(tab []utils.Record) string {

	var resultat string
	for k := 0; k < len(tab); k++ {
		resultat += strconv.Itoa(int(tab[k].Type)) + "," + strconv.Itoa(tab[k].ClockValue) + ";"
	}
	return resultat
}

// Enregistrement de l'instantanée sous forme de fichier texte.
func writeSnapshotInFile() {
	file, err := os.Create("Snapshot.txt")
	if err != nil {
		log.Fatal("Impossible de creer le fichier", err)
	}
	defer file.Close()

	for i := 0; i < nbSite; i++ {
		_, err := fmt.Fprintf(file, "Site %d :  Stock local = %d stock global = %d  tableau horloges = %s\n", i+1, localStocks[i], globalStocks[i], printEG(EG[i]))
		if err != nil {
			log.Fatal("Impossible d'ecrire dans le fichier", err)
		}
	}

	for i := 0; i < nbSite; i++ {
		_, err := fmt.Fprintf(file, "Messages Prepost du Site %d :  %s\n", i+1, prepost[i])
		if err != nil {
			log.Fatal("Impossible d'ecrire dans le fichier", err)
		}
	}
}

// True si le site peut acceder à la section critique, false sinon.
func canEnterCriticalSection() bool {
	if Tab[siteId-1].Type != utils.Request {
		return false
	}

	// Verifie si le site a l'estampille la plus basse.
	for k := 1; k <= nbSite; k++ {
		isSmallestEstampille := ((Tab[k-1].ClockValue) == (Tab[siteId-1].ClockValue) && (k < siteId)) || ((Tab[k-1].ClockValue) < (Tab[siteId-1].ClockValue))
		if k != siteId && (isSmallestEstampille) {
			return false
		}
	}

	return true
}

// True si le message reçu est à envoyer sur l'anneau, false sinon.
func mustForward(message utils.Message) bool {
	return message.Sender != siteId && message.Receiver != siteId
}

// Traitement d'une requête de section critique de l'application de base du site.
func handleSCRequest() {
	horloge++
	Tab[siteId-1] = utils.Record{utils.Request, horloge}
	utils.SendAll(utils.Request, siteId, horloge, -1)
	bilan = bilan + nbSite - 1
}

// Traitement d'une libération de section critique par l'application de base du site.
func handleSCRelease(stock int) {
	horloge++
	Tab[siteId-1] = utils.Record{utils.Release, horloge}
	utils.SendAll(utils.Release, siteId, horloge, stock)
}

// Traitement d'une requête de section critique par un autre site.
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

// Traite la demande de lancement d'un instantané du système.
func startSnapShot(globalStock int, localStock int) {
	couleur = utils.Rouge
	initiateur = true
	EG[siteId-1] = Tab
	NbEtatsAttendus = nbSite - 1
	NbMessagesAttendus = bilan
	globalStocks[siteId] = globalStock
	localStocks[siteId] = localStock
	bilan++
}

// Traite la réception d'un message d'état d'un autre site.
func handleEtat(sender int, etat []utils.Record, bilan int) {
	NbMessagesAttendus += bilan
	EG[sender-1] = etat
	NbEtatsAttendus--

	if NbMessagesAttendus == 0 && NbEtatsAttendus == 0 {
		writeSnapshotInFile()
	}
}

// Traite la réception d'un message prépost d'un autre site.
func handlePrepost(sender int, prepostMessage string) {
	NbMessagesAttendus--
	prepost[sender] += prepostMessage + "\n"

	if NbMessagesAttendus == 0 && NbEtatsAttendus == 0 {
		writeSnapshotInFile()
	}
}

// Traite la demande d'information de stock émise par l'application de base
func handleStockRequest(globalStock int, localStock int) {
	//CLOCK = LOCALSTOCK
	utils.Send(utils.Etat, siteId, siteId, localStock, globalStock, couleur, Tab, bilan, "")
}

// Traiter le message en fonction de son type
func handleMessage(message utils.Message) {
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
		// CLOCK = LOCAL STOCK
		startSnapShot(message.GlobalStock, message.ClockValue)
	case utils.Etat:
		handleEtat(message.Sender, message.Etat, message.Bilan)
	case utils.Prepost:
		handlePrepost(message.Sender, message.PrepostMessage)
	case utils.StockRequest:
		// CLOCK = LOCAL STOCK
		handleStockRequest(message.GlobalStock, message.ClockValue)
	}
}

// Routine attendant la reception d'un message et lançant le traitement du message
func waitMessages() {
	for {
		message := utils.Receive()

		if message.Couleur == utils.Blanc && couleur == utils.Rouge {
			utils.Send(utils.SCStart, siteId, siteId, horloge, -1, couleur, Tab, bilan, utils.EncodeSimpleMessage(message))
		}

		if message.Couleur == utils.Rouge && couleur == utils.Blanc {
			couleur = utils.Rouge
			utils.Send(utils.StockRequest, siteId, siteId, horloge, -1, couleur, Tab, bilan, "")
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

	EG = make([][]utils.Record, nbSite)
	for k := 0; k < nbSite; k++ {
		EG[k] = make([]utils.Record, nbSite)
	}

	globalStocks = make([]int, nbSite)
	localStocks = make([]int, nbSite)
	prepost = make([]string, nbSite)

	// Initialiser l'horloge
	horloge = 0

	// Lancement de la routine
	go waitMessages()

	// Attente infinie.
	select {}
}
