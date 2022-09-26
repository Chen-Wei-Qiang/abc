package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/cloudmailin/cloudmailin-go"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		message, err := cloudmailin.ParseIncoming(req.Body)
		if err != nil {
			http.Error(w, "Error parsing message: "+err.Error(), http.StatusUnprocessableEntity)
			return
		}

		if strings.HasPrefix(message.Envelope.To, "noreply@") {
			http.Error(w, "No replies please", http.StatusForbidden)
			return
		}

		body := message.ReplyPlain
		if body == "" {
			body = message.Plain
		}

		fmt.Fprintln(w, "Thanks for message: ", message.Headers.First("message_id"))

		log.Println("Reply: ", body)
		log.Println("HTML: ", message.HTML)

		log.Println("Attachment Name: ", message.Attachments[0].FileName)
		log.Println("Attachment URL: ", message.Attachments[0].URL)
	})

	http.ListenAndServe("imap.qq.com:993", nil)
}
