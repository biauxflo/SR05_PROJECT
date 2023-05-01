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
	Prepost
	Etat
	SnapStart
	SnapInfo
)

type Message struct {
	Type        MessageType
	Sender      int
	Receiver    int
	ClockValue  int
	GlobalStock int
	Color 			Couleur
	LocalStock  int
	Bilan 			int
	Tab  				string
}

type PrepostMessage struct {
	Type           MessageType
	Sender         int
	Receiver       int
	InitialMessage Message
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
	clock := msg_format("ClockValue", strconv.Itoa(msg.ClockValue))
	globalStock := msg_format("GlobalStock", strconv.Itoa(msg.GlobalStock))
	color := msg_format("Color", strconv.Itoa(int(msg.Color)))
	localstock :=  msg_format("LocalStock", strconv.Itoa(msg.LocalStock))
	bilan := msg_format("Bilan", strconv.Itoa(msg.Bilan))
	tab :=  msg_format("Tab", msg.Tab)
	return msgType + sender + receiver + clock + globalStock + color + localstock + bilan + tab
}

func EncodePrepost(prep PrepostMessage) string {
	msgType := msg_format("IsPrep", strconv.Itoa(int(prep.Type)))
	sender := msg_format("PrepSender", strconv.Itoa(prep.Sender))
	receiver := msg_format("PrepReceiver", strconv.Itoa(prep.Receiver))
	message := EncodeMessage(prep.InitialMessage)
	return msgType + sender + receiver + message
}

func msg_send(msg string) {
	mutex.Lock()
	fmt.Println(msg)
	mutex.Unlock()
}

func Send(msgType MessageType, sender int, receiver int, clockValue int, globalStock int, color Couleur, localstock int, bilan int, tab string) {
	message := Message{Type: msgType, Sender: sender, ClockValue: clockValue, Receiver: receiver, GlobalStock: globalStock, Color: color, LocalStock: localstock, Bilan: bilan, Tab: tab}
	l := log.New(os.Stderr, "", 0)
	l.Println(strconv.Itoa(sender) + " --> " + EncodeMessage(message))
	msg_send(EncodeMessage(message))
}

func SendAll(msgType MessageType, sender int, clockValue int, globalStock int, color Couleur, localstock int, bilan int, tab string) {
	Send(msgType, sender, 0, clockValue, globalStock, color, localstock, bilan, tab)
}

func SendPrePost(sender int, receiver int, message Message) {
	newmessage := PrepostMessage{Type: Prepost, Sender: sender, Receiver: receiver, InitialMessage: message}
	k := log.New(os.Stderr, "", 0)
	k.Println(strconv.Itoa(sender) + " --> " + EncodePrepost(newmessage))
	msg_send(EncodeMessage(message))
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

func Receive() (Message, PrepostMessage) {
	var rcvmsg, msgType, sender, clockValue, receiver, globalStock, color, localstock, bilan, tab string
	var msgTypeRcv int
	var senderRcv, clockValueRcv, receiverRcv, globalStockRcv, colorRcv, localstockRcv, bilanRcv int
	var prep PrepostMessage

	fmt.Scanln(&rcvmsg)
	mutex.Lock()

	msgType = Findval(rcvmsg, "Type")
	sender = Findval(rcvmsg, "Sender")
	clockValue = Findval(rcvmsg, "ClockValue")
	sender = Findval(rcvmsg, "Sender")
	globalStock = Findval(rcvmsg, "GlobalStock")
	color = Findval(rcvmsg, "Color")
	localstock = Findval(rcvmsg, "LocalStock")
	bilan = Findval(rcvmsg,"Bilan")
	tab = Findval(rcvmsg, "Tab")
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
	if localstock != "" {
		localstockRcv, _ = strconv.Atoi(localstock)
	}
	if bilan != "" {
		bilanRcv, _ = strconv.Atoi(bilan)
	}

	mutex.Unlock()
	rcvmsg = ""
	message := Message{Type: MessageType(msgTypeRcv), Sender: senderRcv, Receiver: receiverRcv, ClockValue: clockValueRcv, GlobalStock: globalStockRcv, Color: Couleur(colorRcv), LocalStock: localstockRcv, Bilan: bilanRcv, Tab: tab}

	isPrep := Findval(rcvmsg, "IsPrep")
	if isPrep != "" {
		prep.Type = Prepost

		prepSender := Findval(rcvmsg, "Sender")
		prepReceiver := Findval(rcvmsg, "Receiver")

		prep.Receiver, _ = strconv.Atoi(prepReceiver)
		prep.Sender, _ = strconv.Atoi(prepSender)

		prep.InitialMessage = message
	}
	return message, prep
}

func Forward(message Message) {
	msg_send(EncodeMessage(message))
}

func ForwardPrepost(message PrepostMessage) {
	msg_send(EncodePrepost(message))
}
