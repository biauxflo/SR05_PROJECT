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
)

type MessageType int

const (
	Release MessageType = iota
	ACK
	Request
	SCRequest
	SCStart
	SCEnd
	Etat
	PrePost
	SnapStart
)

type Message struct {
	Type        MessageType
	Sender      int
	Receiver    int
	ClockValue  int
	GlobalStock int
	Color 			Couleur
	TypeOfMessage MessageType
}

var fieldsep = "/"
var keyvalsep = "="
var mutex = &sync.Mutex{}

func msg_format(key string, val string) string {
	return fieldsep + key + keyvalsep + val
}

func EncodeMessage(msg Message) string {
	msgType := msg_format("Type", strconv.Itoa(int(msg.Type)))
	sender := msg_format("Sender", strconv.Itoa(msg.Sender))
	receiver := msg_format("Receiver", strconv.Itoa(msg.Receiver))
	clock := msg_format("ClockCount", strconv.Itoa(msg.ClockValue))
	stock := msg_format("GlobalStock", strconv.Itoa(msg.GlobalStock))
	color := msg_format("Color", strconv.Itoa(int(msg.Color)))
	typeOfMessage := msg_format("TypeOfMessage", strconv.Itoa(int(msg.TypeOfMessage)))
	return msgType + sender + receiver + clock + stock + color + typeOfMessage
}

func msg_send(msg string) {
	fmt.Println(msg)
}

func Send(msgType MessageType, sender int, receiver int, clockValue int, globalStock int, color Couleur, typeOfMessage MessageType) {
	message := Message{Type: msgType, Sender: sender, ClockValue: clockValue, Receiver: receiver, GlobalStock: globalStock, Color: color, TypeOfMessage: typeOfMessage}
	l := log.New(os.Stderr, "", 0)
	l.Println(strconv.Itoa(sender) + EncodeMessage(message))
	msg_send(EncodeMessage(message))
}

func SendAll(msgType MessageType, sender int, clockValue int, globalStock int, color Couleur, typeOfMessage MessageType) {
	Send(msgType, sender, 0, clockValue, globalStock, color, typeOfMessage)
}

func Findval(msg string, key string) string {
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

func Receive() Message {
	var rcvmsg, msgType, sender, clockValue, receiver, globalStock, color, typeOfMessage string
	var msgTypeRcv int
	var senderRcv, clockValueRcv, receiverRcv, globalStockRcv, colorRcv, typeOfMessageRcv int

	fmt.Scanln(&rcvmsg)
	mutex.Lock()

	msgType = Findval(rcvmsg, "Type")
	sender = Findval(rcvmsg, "Sender")
	clockValue = Findval(rcvmsg, "ClockValue")
	receiver = Findval(rcvmsg, "Receiver")
	globalStock = Findval(rcvmsg, "GlobalStock")
	color = Findval(rcvmsg, "Color")
	typeOfMessage = Findval(rcvmsg, "TypeOfMessage")
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
	if typeOfMessage != "" {
		typeOfMessageRcv, _ = strconv.Atoi(typeOfMessage)
	}

	mutex.Unlock()
	rcvmsg = ""

	return Message{Type: MessageType(msgTypeRcv), Sender: senderRcv, Receiver: receiverRcv, ClockValue: clockValueRcv, GlobalStock: globalStockRcv, Color: Couleur(colorRcv), TypeOfMessage: MessageType(typeOfMessageRcv)}
}

func Forward(message Message) {
	msg_send(EncodeMessage(message))
}
