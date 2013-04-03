package stream

import (
	"bufio"
	"encoding/json"
	"log"
	"mxml/makini/api"
	"net/http"
	"time"
)

func consumeStream(url string, msgChan chan []byte) {
	for {
		res, err := http.Get(url)

		if err == nil {
			buf := bufio.NewReader(res.Body)

			for {
				line, err := buf.ReadBytes('\n')
				length := len(line)

				if err == nil && length >= 2 {
					line = line[:length-2]
					if len(line) > 0 {
						msgChan <- line
					}
				} else {
					log.Print("Error while reading: ", err)
					break
				}
			}
		} else {
			log.Print("Error while connecting: ", err)
		}

		time.Sleep(time.Second)
	}
}

func unmarshalStream(in chan []byte, out chan *api.APIResponse) {
	for {
		m := &api.APIResponse{}
		msg := <-in
		err := json.Unmarshal(msg, &m)

		if err != nil {
			log.Print("Error decoding: ", msg, err)
		} else {
			out <- m
		}
	}
}

func ProcessStream(url string) chan *api.APIResponse {
	bytes := make(chan []byte)
	go consumeStream(url, bytes)

	messages := make(chan *api.APIResponse)
	go unmarshalStream(bytes, messages)

	return messages
}
