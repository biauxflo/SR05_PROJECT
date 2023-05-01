package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
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
)

type Message struct {
	Type        MessageType
	Sender      int
	Receiver    int
	ClockValue  int
	GlobalStock int
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
	stock := msg_format("GlobalStock", strconv.Itoa(msg.GlobalStock))
	return msgType + sender + receiver + clock + stock
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

func Send(msgType MessageType, sender int, receiver int, clockValue int, globalStock int) {
	message := Message{Type: msgType, Sender: sender, ClockValue: clockValue, Receiver: receiver, GlobalStock: globalStock}
	l := log.New(os.Stderr, "", 0)
	l.Println(strconv.Itoa(sender) + " --> " + EncodeMessage(message))
	msg_send(EncodeMessage(message))
}

func SendAll(msgType MessageType, sender int, clockValue int, globalStock int) {
	Send(msgType, sender, 0, clockValue, globalStock)
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
	var rcvmsg, msgType, sender, clockValue, receiver, globalStock string
	var msgTypeRcv int
	var senderRcv, clockValueRcv, receiverRcv, globalStockRcv int
	var prep PrepostMessage

	fmt.Scanln(&rcvmsg)
	mutex.Lock()

	msgType = Findval(rcvmsg, "Type")
	sender = Findval(rcvmsg, "Sender")
	clockValue = Findval(rcvmsg, "ClockValue")
	sender = Findval(rcvmsg, "Sender")
	globalStock = Findval(rcvmsg, "GlobalStock")
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

	mutex.Unlock()
	rcvmsg = ""
	message := Message{Type: MessageType(msgTypeRcv), Sender: senderRcv, Receiver: receiverRcv, ClockValue: clockValueRcv, GlobalStock: globalStockRcv}

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
