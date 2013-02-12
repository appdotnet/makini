package stream

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"time"
	"makini/api"
)

func ConsumeStream(url string, msgChan chan []byte) {
	for {
		res, err := http.Get(url)

		if err == nil {
			buf := bufio.NewReader(res.Body)

			for {
				line, err := buf.ReadBytes('\n')

				if err == nil {
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

func UnmarshalStream(in chan []byte, out chan *api.APIResponse) {
	for {
		m := &api.APIResponse{}
		err := json.Unmarshal(<-in, &m)

		if err != nil {
			log.Print("Error decoding: ", err)
		} else {
			out <- m
		}
	}
}
