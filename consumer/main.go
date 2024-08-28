package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// connect to rabbitmq
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	checkErr(err)
	defer conn.Close()

	// open a channel (client)
	ch, err := conn.Channel()
	checkErr(err)
	defer ch.Close()

	// create a queue (if already created will do nothing)
	q, err := ch.QueueDeclare(
		"test-queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	checkErr(err)

	// receive messages
	msgs, err := ch.Consume(
		q.Name,
		"consumer1",
		false,
		false,
		false,
		false,
		nil,
	)
	checkErr(err)

	go func() {
		for d := range msgs {
			ex := d.Exchange
			fmt.Printf("len(ex): %v\n", len(ex))
			if len(ex) == 0 {
				ex = "(default)"
			}
			log.Printf("Received: %s from: %s\n", d.Body, ex)
		}
	}()

	log.Printf("[*] Waiting for messages. To exit precc CTRL+C")
	<-make(chan bool)
}
