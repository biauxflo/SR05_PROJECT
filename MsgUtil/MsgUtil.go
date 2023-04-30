package MsgUtil

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
	Type        MessageType
	Sender      int
	Receiver    int
	ClockValue  int
	GlobalStock int
}

var fieldsep = "/"
var keyvalsep = "="

func EncodeMessage(msg Message) string {
	return msg_format("Type", EncodeMessageType(msg.Type)) + msg_format("Sender", strconv.Itoa(msg.Sender)) + msg_format("Receiver", strconv.Itoa(msg.Receiver)) + msg_format("ClockCount", strconv.Itoa(msg.ClockValue)) + msg_format("GlobalStock", strconv.Itoa(msg.GlobalStock))
}

func EncodeMessageType(msg MessageType) string {
	res := ""
	switch msg {
	case Release:
		res = "release"
	case ACK:
		res = "ack"
	case Request:
		res = "request"
	case SCRequest:
		res = "screquest"
	case SCStart:
		res = "scstart"
	case SCEnd:
		res = "scend"
	}

	return res
}

func DecodeMessageType(msgType string) (MessageType, error) {
	switch strings.ToLower(msgType) {
	case "release":
		return Release, nil
	case "ack":
		return ACK, nil
	case "request":
		return Request, nil
	case "screquest":
		return SCRequest, nil
	case "scstart":
		return SCStart, nil
	case "scend":
		return SCEnd, nil
	default:
		return -1, errors.New("Type inconnu : " + msgType)
	}
}

func Send(sender int, clockValue int, receiver int, globalStock int) {
	msg_send(EncodeMessage(Message{Type: Request, Sender: sender, ClockValue: clockValue, Receiver: receiver, GlobalStock: globalStock}))
}

func SendAll(sender int, clockValue int) {
	msg_send(EncodeMessage(Message{Type: Release, Sender: sender, ClockValue: clockValue, Receiver: 0}))
}

func SendAllRelease(sender int, clockValue int, globalStock int) {
	msg_send(EncodeMessage(Message{Type: Release, Sender: sender, ClockValue: clockValue, Receiver: 0, GlobalStock: globalStock}))
}

func msg_format(key string, val string) string {
	return fieldsep + key + keyvalsep + val
}

func msg_send(msg string) {
	fmt.Print(msg + "\n")
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

var mutex = &sync.Mutex{}

func Receive() Message {
	var rcvmsg, msgType, sender, clockValue, receiver, globalStock string
	var msgTypeRcv MessageType
	var senderRcv, clockValueRcv, receiverRcv, globalStockRcv int
	l := log.New(os.Stderr, "", 0)

	fmt.Scanln(&rcvmsg)
	mutex.Lock()
	l.Println("reception <", rcvmsg, ">")
	msgType = Findval(rcvmsg, "Type")
	sender = Findval(rcvmsg, "Sender")
	clockValue = Findval(rcvmsg, "ClockValue")
	receiver = Findval(rcvmsg, "Receiver")
	globalStock = Findval(rcvmsg, "GlobalStock")
	if msgType != "" {
		msgTypeRcv, _ = DecodeMessageType(msgType)
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
	for i := 1; i < 6; i++ {
		l.Println("traitement message", i)
		time.Sleep(time.Duration(1) * time.Second)
	}
	mutex.Unlock()
	rcvmsg = ""

	return Message{Type: msgTypeRcv, Sender: senderRcv, Receiver: receiverRcv, ClockValue: clockValueRcv, GlobalStock: globalStockRcv}
}
