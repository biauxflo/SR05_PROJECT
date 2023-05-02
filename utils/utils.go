package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Couleur int

const (
	Blanc Couleur = iota
	Rouge
	Neutre
)

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

type Record struct {
	Type       MessageType
	ClockValue int
}

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

func printVec(tab []Record) string {
	var resultat string
	for k := 0; k < len(tab); k++ {
		resultat += strconv.Itoa(int(tab[k].Type)) + " " + strconv.Itoa(tab[k].ClockValue)
	}
	return resultat
}
func parseVec(vec string) []Record {
	var tab []Record
	if len(vec) >= 3 {
		tab = make([]Record, len(vec)/3)
		for k := 0; k < len(vec); k += 3 {
			if k == 0 {
				tab[k] = Record{MessageType(int(vec[k] - '0')), int(vec[k+2] - '0')}
			} else {
				tab[k/3] = Record{MessageType(int(vec[k] - '0')), int(vec[k+2] - '0')}
			}
		}
		return tab
	}
	return nil
}

func msg_format(key string, val string) string {
	return fieldsep + key + keyvalsep + val
}

func prepost_format(val string) string {
	return prepostSeparator + "PrepostMessage" + keyvalsep + val + prepostSeparator
}

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

func EncodeMessage(msg Message) string {
	encodedMsg := EncodeSimpleMessage(msg)
	prepost := prepost_format(msg.PrepostMessage)
	return encodedMsg + prepost
}

func msg_send(msg string) {
	fmt.Println(msg)
}

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
	l := log.New(os.Stderr, "", 0)
	l.Println(strconv.Itoa(sender) + " --> " + EncodeMessage(message))
	msg_send(EncodeMessage(message))
}

func SendAll(msgType MessageType, sender int, clockValue int, globalStock int) {
	Send(msgType, sender, 0, clockValue, globalStock, 2, nil, 0, "")
}

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

func Receive() Message {
	var rcvmsg string
	var senderRcv, clockValueRcv, receiverRcv, globalStockRcv, msgTypeRcv, colorRcv, bilanRcv int
	var etatRcv []Record

	fmt.Scanln(&rcvmsg)

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

func Forward(message Message) {
	msg_send(EncodeMessage(message))
}
