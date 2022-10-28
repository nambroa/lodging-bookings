package main

import (
	"github.com/nambroa/lodging-bookings/internal/models"
	mail "github.com/xhit/go-simple-mail/v2"
	"log"
	"runtime/debug"
	"time"
)

func listenForMail() {
	// Function that executes on the background indefinitely.
	go func() {
		for {
			msg := <-app.Mailchan // Listen for mail on the mail channel.
			sendMsg(msg)
		}
	}()
}

func sendMsg(m models.MailData) {
	// Create client and mail server.
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false // Only make connection to mail server when sending an email.
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		log.Println("Error connecting to mailserver!")
		log.Println(err)
		debug.PrintStack()
		//errorLog.Println(err)
	}

	// Build email message.
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	email.SetBody(mail.TextHTML, m.Content)

	err = email.Send(client)
	if err != nil {
		log.Println("Error sending mail!")
		log.Println(err)
		debug.PrintStack()
	} else {
		log.Printf("Mail sent successfully from %s to %s", m.From, m.To)
	}
}
