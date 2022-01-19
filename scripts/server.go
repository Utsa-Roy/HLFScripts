/*Mentor: Utsa Roy
Developer: Riom Sen
Version: 0.02
©IIEST, Shibpur*/

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"example.com/cryptofunc"
	"github.com/dustin/go-coap"
)

var ch chan int
var client_id [6]byte
var iot_id [6]byte
var resource_id [6]byte
var key [32]byte
var message_id int
var response []byte

func monitor1() {
	/*This function is the most unique of the all the functions present here. It
	  demonstrates the native concurrency capability of Go. First of all , ch is a
	  channel, a sibling of Unix pipe. We can send and receive values across threads
	  using it. One interesting property of channel is that reading an empty channel
	  or writing to a full channel blocks the application untill the operation is
	  executed.

	  This is used in select case construct which executes the blocking case which
	  executes the first. The infinite for loop ensures that after every client
	  request is received, the timer is reset. This function is executed parallely in
	  a separate thread.
	*/
	for {
		select {
		case <-ch:
			log.Println("Serving a client request")
		case <-time.After(120 * time.Second): //if no client request for more than 2 minutes
			log.Println("No client request for a long time.")
			os.Exit(0)
		}
	}
}
func monitor2() {
	select {
	case <-time.After(600 * time.Second): //stop the server after every 10 minutes anyway
		log.Println("Shutting down server anyway.")
		os.Exit(0)
	}
}

func handleError(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
		os.Exit(1)
	}
}

func authenticate(client_id []byte, iot_id []byte, resource_id []byte) bool {
	start := time.Now()

	cmd, err := exec.Command("/bin/sh", "R1.sh").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output := string(cmd)
	fmt.Println(output)

	cmd1, err := exec.Command("/bin/sh", "qr1.sh").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output1 := string(cmd1)
	fmt.Println(output1)

	elapsed := time.Since(start)
	fmt.Printf("Time taken %s", elapsed)
	return true
	//This function will contain the authentication logic
}
func handleA(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	ch <- 1
	clientpublic, err := cryptofunc.ExtractPublicKey("Client.publickey")
	handleError("Unable to extract public key: %s ", err)
	serverprivate, err := cryptofunc.ExtractPrivateKey("Server.privatekey")
	handleError("Unable to extract private key: %s ", err)

	//log.Printf("Got message in handleA:\n path=%q:\n %#v from %v\n\n", m.Path(), m, a)
	//The required keys were collected
	lclient_id, liot_id, lresource_id, sym_key, err := cryptofunc.UnpackSOR(m.Payload, clientpublic, serverprivate)
	handleError("Unpacking failed: %s", err)
	authentic := authenticate(lclient_id, liot_id, lresource_id)
	if authentic {
		//store credentials and inform client
		copy(client_id[:], lclient_id[:])
		copy(iot_id[:], liot_id[:])
		copy(resource_id[:], lresource_id[:])
		copy(key[:], sym_key)
		response, err = cryptofunc.Encrypt(key, []byte("SUCCESS"))
		log.Printf("Got a new rquest.\nClient id: %s IOT id: %s Resource id: %s\n", client_id, iot_id, resource_id)
	} else {
		response = []byte("FAILURE")
	}

	if m.IsConfirmable() {
		//generating reply
		res := &coap.Message{
			Type:      coap.Acknowledgement,
			Code:      coap.Content,
			MessageID: m.MessageID,
			Token:     m.Token,
			Payload:   response,
		}
		res.SetOption(coap.ContentFormat, coap.TextPlain)

		//log.Printf("Transmitting from A %#v", res)
		return res
	}
	return nil
}

func handleB(l *net.UDPConn, a *net.UDPAddr, m *coap.Message) *coap.Message {
	//does job based on stored credential
	//log.Printf("Got message in handleB: path=%q: %#v from %v", m.Path(), m, a)
	ch <- 1
	msg, err := cryptofunc.Decrypt(key, m.Payload)
	handleError("Decryption failed %s \n", err)
	log.Printf("Message received is %s \n", msg)
	response, err := cryptofunc.Encrypt(key, []byte("PONG"))
	//Sample job. Must be replaced by something meaningful.
	handleError("Encryption failed %s \n", err)
	if m.IsConfirmable() {
		res := &coap.Message{
			Type:      coap.Acknowledgement,
			Code:      coap.Content,
			MessageID: m.MessageID,
			Token:     m.Token,
			Payload:   response,
		}
		res.SetOption(coap.ContentFormat, coap.TextPlain)

		//log.Printf("Transmitting from B %#v", res)
		return res
	}
	return nil
}

func main1() {
	ch = make(chan int)
	mux := coap.NewServeMux()
	mux.Handle("/authenticate", coap.FuncHandler(handleA))
	mux.Handle("/execute", coap.FuncHandler(handleB))
	log.Printf("SERVER STARTED\n")
	go monitor1() //execute this function in a separate thread
	go monitor2()
	log.Fatal(coap.ListenAndServe("udp", ":5683", mux))
}
