package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"utils"
)

type Record struct {
	Type       utils.MessageType
	ClockValue int
}

var Tab []Record
var EG []string
var GlobalStocks []int
var LocalStocks []int
var PrePost []utils.PrepostMessage
var horloge int
var nbSite int
var siteId int
var couleur utils.Couleur
var bilan int
var nbEtatsAttendus int
var nbMsgAttendus int

func printTab() {
	l := log.New(os.Stderr, "", 0)
	l.Println(strconv.Itoa(siteId) + ": ")
	for k := 0; k < nbSite; k++ {
		l.Print(Tab[k])
	}
}

func printVec(tab []Record) string {
	var resultat string
	for k := 0; k < nbSite; k++ {
		 resultat += strconv.Itoa(int(tab[k].Type)) + "," + strconv.Itoa(tab[k].ClockValue)
	}
	return resultat
}

func writeSnapshotInFile() {
    file, err := os.Create("Snapshot.txt")
    if err != nil {
        log.Fatal("Impossible de creer le fichier", err)
    }
    defer file.Close()

    for i := 0; i < nbSite; i++ {
        _, err := fmt.Fprintf(file, "Site %d :  Stock local = %d stock global = %d  tableau horloges = %s\n", i+1, LocalStocks[i], GlobalStocks[i], EG[i])
        if err != nil {
            log.Fatal("Impossible d'ecrire dans le fichier", err)
        }
    }
		count :=0
		for _, msg := range PrePost {
			count++
        _, err := file.WriteString(fmt.Sprintf("Message PrePost numero : %d : [ Type: %d, Sender: %d, Receiver: %d, ClockValue: %d GlobalStock: %d Color: %d	LocalStock: %d	Bilan: %d	Tab: %s ]\n",count, msg.InitialMessage.Type, msg.InitialMessage.Sender, msg.InitialMessage.Receiver, msg.InitialMessage.ClockValue, msg.InitialMessage.GlobalStock, msg.InitialMessage.Color, msg.InitialMessage.LocalStock, msg.InitialMessage.Bilan, msg.InitialMessage.Tab))
        if err != nil {
            log.Fatal("Cannot write to file", err)
        }
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

func handleSCRequest(color utils.Couleur) {
	horloge++
	Tab[siteId-1] = Record{utils.Request, horloge}
	bilan++
	utils.SendAll(utils.Request, siteId, horloge, -1, color, -1, -1, "")
}

func handleSCRelease(stock int, color utils.Couleur) {
	horloge++
	Tab[siteId-1] = Record{utils.Release, horloge}
	bilan++
	utils.SendAll(utils.Release, siteId, horloge, stock, color, -1, -1, "")
}

func handleRequest(h int, sender int, color utils.Couleur) {
	horloge = max(horloge, h) + 1
	Tab[sender-1] = Record{utils.Request, h}
	bilan++
	utils.Send(utils.ACK, siteId, sender, horloge, -1, color, -1, -1, "")
	if canEnterCriticalSection() {
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, color, -1, -1, "")
	}
}

// Fonction pour gérer la réception d'un message de libération
func handleRelease(h int, sender int, globalStock int, color utils.Couleur) {
	// Mettre à jour la date logique du site local
	horloge = max(horloge, h) + 1

	// Mettre à jour le tableau Tab
	Tab[sender-1] = Record{utils.Release, h}

	bilan--
	// Mettre à jour la valeur du stock dans l'application
	utils.Send(utils.SCUpdate, siteId, siteId, horloge, globalStock, color, -1, -1, "")

	// Vérifier si la condition pour entrer en section critique est satisfaite
	if canEnterCriticalSection() {
		bilan--
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, color, -1, -1, "")
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
		bilan--
		// Si oui, envoyer un message à l'application de base pour commencer la section critique
		utils.Send(utils.SCStart, siteId, siteId, horloge, -1, color, -1, -1, "")
	}
}

// Fonction pour gérer la réception d'un message de début de snapshot
func handleSnapStart(globalStock int, localStock int) {
	//si une snapshot n'est pas en cours et que l'on est bien le site 1 qui s'occupe des snapshots
	if (siteId == 1){
		//on change de couleur pour signaliser une nouvelle snapshot
		couleur = utils.Rouge

		//on stock l'état local
		EG[siteId-1] = printVec(Tab)
		GlobalStocks[siteId-1] = globalStock
		LocalStocks[siteId-1] = localStock

		nbEtatsAttendus = nbSite - 1
		nbMsgAttendus = bilan
	}
}

// Fonction pour gérer la réception d'un message de reception des infos de l'appli
func handleSnapInfo(globalstock int, localstock int) {
	stringTab := printVec(Tab)
	utils.Send(utils.Etat, siteId, 1, -1, globalstock, couleur, localstock, bilan, stringTab)
}

func handleEtat(sender int, bilan int, localStock int, globalStock int, tab string) {
	if siteId == 1{
		//si on est le site 1 alors on le traite sinon on ne fait que de le Forward
		nbEtatsAttendus--
		nbMsgAttendus += bilan

		//on sauvegarde l'état fournie
		LocalStocks[sender-1] = localStock
		GlobalStocks[sender-1] = globalStock
		EG[sender-1] = tab

		if nbEtatsAttendus == 0 && nbMsgAttendus == 0{
			//fin du programme on écrit la snapshot dans le fichier
			writeSnapshotInFile()
		}
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
		handleRelease(message.ClockValue,message.Sender,message.GlobalStock,message.Color)
	case utils.ACK:
		handleAck(message.ClockValue, message.Sender,message.Color)
	case utils.Etat:
		handleEtat(message.Sender, message.Bilan, message.LocalStock, message.GlobalStock, message.Tab)
	case utils.SnapStart:
		handleSnapStart(message.GlobalStock, message.LocalStock)
	case utils.SnapInfo:
		handleSnapInfo(message.GlobalStock,message.LocalStock)
	}
}

func waitMessages() {
	for {
		message, prep := utils.Receive()

		//handlePrePost
		if prep.Type == utils.Prepost {
			if siteId == 1{
				//on ajoute le messsage prépost à la liste de sauvegarde
				PrePost = append(PrePost, prep)
				nbMsgAttendus--
				if nbEtatsAttendus == 0 && nbMsgAttendus == 0{
					//fin du programme on écrit la snapshot dans le fichier
					writeSnapshotInFile()
				}
			}else{
				//on forward si on est pas le site 1
				utils.ForwardPrepost(prep)
			}
		} else {
			l := log.New(os.Stderr, "", 0)
			l.Println(strconv.Itoa(siteId) + " <-- Type : " + strconv.Itoa(int(message.Type)) + " Sender :  " + strconv.Itoa(message.Sender) + " Clock : " + strconv.Itoa(message.ClockValue))
			if (message.Sender == siteId){
				//on réduit le bilan dans le cas ou le message a parcouru tout l'anneau
				bilan--
			}
			if message.Color != utils.Rouge && couleur == utils.Blanc{
				//première réception de couleur rouge il faut donc prendre une snapshot
				couleur = utils.Rouge
				//on envoi un message de type info pour demander les info à l'appli avant de l'envoyer au site 1
				utils.Send(utils.SnapInfo, siteId, siteId, horloge, -1, couleur, -1, -1, "")
			}
			if message.Color == utils.Blanc && couleur == utils.Rouge{
				//cas message est un prépost
				utils.SendPrePost(siteId,1,message)
			}
			if (message.Receiver == 0 && message.Sender != siteId) || message.Receiver == siteId {
				handleMessage(message)
			}
			if mustForward(message) {
				utils.Forward(message)
			}
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
	EG = make([]string, nbSite)
	GlobalStocks = make([]int, nbSite)
	for k := 0; k < nbSite; k++ {
		Tab[k] = Record{utils.Release, 0}
	}

	// Initialiser l'horloge
	horloge = 0
	//initialiser la couleur du site
	couleur = utils.Blanc
	//initialisations pour la snapshot
	bilan = 0
	nbMsgAttendus = 0
	nbEtatsAttendus = 0


	go waitMessages()

	select {}
}
