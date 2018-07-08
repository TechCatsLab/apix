/*
 * Revision History:
 *     Initial: 2018/07/08        Wang RiYu
 */

package main

import (
	"bufio"
	"log"
	"os"

	"github.com/TechCatsLab/apix/nsq"
)

func main() {
	addr := "127.0.0.1:4150"
	lookupdHTTPAddress := "127.0.0.1:4161"

	c1, err := nsq.NewClient(1, addr, lookupdHTTPAddress, nil)
	if err != nil {
		log.Fatalf("%#v\n", err)
	}

	c1.Subscribe("c2-c1", "channel_2", nil, nsq.HandlerFunc(func(msg *nsq.Message) error {
		log.Println("receive", msg.NSQDAddress, "message:", string(msg.Body))
		return nil
	}))

	reader := bufio.NewReader(os.Stdin)
	for {
		data, _, _ := reader.ReadLine()
		message := string(data)
		if message == "stop" {
			break
		}

		log.Printf("send message: %s\n", message)
		if err := c1.Publish("c1-c2", message); err != nil {
			log.Fatalf("%#v\n", err)
			break
		}
	}

	c1.Producer.Stop()
}
