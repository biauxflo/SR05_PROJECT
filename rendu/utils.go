package utils

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Couleur utilisée dans l'alorithme de prise d'instantané.
type Couleur int

const (
	Blanc Couleur = iota
	Rouge
	Neutre
)

// MessageType Type de messages émis et reçus au sein de l'application
type MessageType int

const (
	Release MessageType = iota
	ACK
	Request
	SCRequest
	SCStart
	SCEnd
	SCUpdate
	SnapStart
	Prepost
	StockRequest
	Etat
)

// Estampille
type Record struct {
	Type       MessageType
	ClockValue int
}

// Message émis dans l'application
type Message struct {
	Type        MessageType
	Sender      int
	Receiver    int
	ClockValue  int
	GlobalStock int
	Couleur     Couleur
	//Pour les messages de snapshot uniquement
	Etat           []Record
	Bilan          int
	PrepostMessage string
}

var fieldsep = "/"
var keyvalsep = "="
var prepostSeparator = "@"
var mutex = &sync.Mutex{}

// Renvoie le tableau d'estampille sous forme de string.
func printVec(tab []Record) string {
	var resultat string
	for k := 0; k < len(tab); k++ {
		resultat += strconv.Itoa(int(tab[k].Type)) + "," + strconv.Itoa(tab[k].ClockValue) + ";"
	}
	return resultat
}

// Parse le tableau d'estampille depuis une string.
func parseVec(vec string) []Record {
	var tab []Record
	div := strings.Count(vec, ";")
	if div >= 1 {
		splitted := strings.Split(vec, ";")
		tab = make([]Record, div)
		for k := 0; k < div; k++ {
			newSplit := strings.Split(splitted[k], ",")
			msgType, _ := strconv.Atoi(newSplit[0])
			clock, _ := strconv.Atoi(newSplit[1])
			tab[k] = Record{MessageType(msgType), clock}
		}
		return tab
	}
	return nil
}

// Formatte la string d'un champ du message
func msg_format(key string, val string) string {
	return fieldsep + key + keyvalsep + val
}

// Formatte le champ contenant le message prepost d'un message
func prepost_format(val string) string {
	return prepostSeparator + "PrepostMessage" + keyvalsep + val + prepostSeparator
}

// Encode un message sans le champ prepost
func EncodeSimpleMessage(msg Message) string {
	msgType := msg_format("Type", strconv.Itoa(int(msg.Type)))
	sender := msg_format("Sender", strconv.Itoa(msg.Sender))
	receiver := msg_format("Receiver", strconv.Itoa(msg.Receiver))
	clock := msg_format("ClockValue", strconv.Itoa(msg.ClockValue))
	stock := msg_format("GlobalStock", strconv.Itoa(msg.GlobalStock))
	color := msg_format("Color", strconv.Itoa(int(msg.Couleur)))
	etat := msg_format("Etat", printVec(msg.Etat))
	bilan := msg_format("Bilan", strconv.Itoa(msg.Bilan))
	return msgType + sender + receiver + clock + stock + color + etat + bilan
}

// Encode un message complet
func EncodeMessage(msg Message) string {
	encodedMsg := EncodeSimpleMessage(msg)
	prepost := prepost_format(msg.PrepostMessage)
	return encodedMsg + prepost
}

// Effectue l'ecriture du message dans le pipe
func msg_send(msg string) {
	fmt.Println(msg)
}

// Envoi un message
func Send(msgType MessageType, sender int, receiver int, clockValue int, globalStock int, couleur Couleur, etat []Record, bilan int, prepost string) {
	message := Message{
		Type:           msgType,
		Sender:         sender,
		ClockValue:     clockValue,
		Receiver:       receiver,
		GlobalStock:    globalStock,
		Couleur:        couleur,
		Etat:           etat,
		PrepostMessage: prepost,
		Bilan:          bilan,
	}
	msg_send(EncodeMessage(message))
}

// Envoi un message à tous les sites de l'application
func SendAll(msgType MessageType, sender int, clockValue int, globalStock int) {
	Send(msgType, sender, 0, clockValue, globalStock, 2, nil, 0, "")
}

// Parse une valeur spécifique au sein d'une chaine de texte formatée.
func findVal(msg string, key string) string {
	if len(msg) < len(fieldsep+key+keyvalsep) {
		return ""
	}

	sep := msg[0:len(fieldsep)]
	tab_allkeyvals := strings.Split(msg[len(fieldsep):], sep)

	for _, keyval := range tab_allkeyvals {
		if len(keyval) >= len(key+keyvalsep) {
			tabkeyval := strings.Split(keyval, keyvalsep)
			if tabkeyval[0] == key {
				return tabkeyval[1]
			}
		}
	}

	return ""
}

// Parse le message prepost contenu dans une chaine de message formattée
func findPrepost(msg string) string {
	if len(msg) < len(prepostSeparator+"PrepostMessage"+keyvalsep) {
		return ""
	}

	sep := msg[0:len(prepostSeparator)]
	tab_allkeyvals := strings.Split(msg[len(prepostSeparator):], sep)

	for _, keyval := range tab_allkeyvals {
		if len(keyval) >= len("PrepostMessage"+keyvalsep) {
			tabkeyval := strings.Split(keyval, keyvalsep)
			if tabkeyval[0] == "PrepostMessage" {
				return tabkeyval[1]
			}
		}
	}

	return ""
}

// Gère la reception et le parsing d'un message dans le pipe
func Receive() Message {
	var rcvmsg string
	var senderRcv, clockValueRcv, receiverRcv, globalStockRcv, msgTypeRcv, colorRcv, bilanRcv int
	var etatRcv []Record

	// Reception
	fmt.Scanln(&rcvmsg)

	mutex.Lock()

	// Parsing
	msgType := findVal(rcvmsg, "Type")
	sender := findVal(rcvmsg, "Sender")
	clockValue := findVal(rcvmsg, "ClockValue")
	receiver := findVal(rcvmsg, "Receiver")
	globalStock := findVal(rcvmsg, "GlobalStock")
	color := findVal(rcvmsg, "Color")
	etat := findVal(rcvmsg, "Etat")
	bilan := findVal(rcvmsg, "Bilan")
	prepost := findPrepost(rcvmsg)
	if msgType != "" {
		msgTypeRcv, _ = strconv.Atoi(msgType)
	}
	if sender != "" {
		senderRcv, _ = strconv.Atoi(sender)
	}
	if clockValue != "" {
		clockValueRcv, _ = strconv.Atoi(clockValue)
	}
	if receiver != "" {
		receiverRcv, _ = strconv.Atoi(receiver)
	}
	if globalStock != "" {
		globalStockRcv, _ = strconv.Atoi(globalStock)
	}
	if color != "" {
		colorRcv, _ = strconv.Atoi(color)
	}
	if etat != "" {
		etatRcv = parseVec(etat)
	}
	if bilan != "" {
		bilanRcv, _ = strconv.Atoi(bilan)
	}

	mutex.Unlock()
	rcvmsg = ""

	message := Message{
		Type:           MessageType(msgTypeRcv),
		Sender:         senderRcv,
		Receiver:       receiverRcv,
		ClockValue:     clockValueRcv,
		GlobalStock:    globalStockRcv,
		Couleur:        Couleur(colorRcv),
		Etat:           etatRcv,
		Bilan:          bilanRcv,
		PrepostMessage: prepost,
	}
	return message
}

// Effectue le forwarding d'un message sur l'anneau
func Forward(message Message) {
	msg_send(EncodeMessage(message))
}
