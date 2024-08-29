package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func consumer() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	checkErr(err)
	defer conn.Close()

	// open a channel
	ch, err := conn.Channel()
	checkErr(err)
	defer ch.Close()

	// open a queue
	queue, err := ch.QueueDeclare(
		"racing",
		true,
		false,
		false,
		false,
		nil,
	)
	checkErr(err)

	err = ch.Qos(1, 0, false)
	checkErr(err)

	// receive messages
	msgs, err := ch.Consume(
		queue.Name,
		"consumer1",
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	checkErr(err)

	go func() {
		for d := range msgs {
			waitFor := rand.Intn(5) + 1

			fmt.Printf("Processing \"%s\", will take %v seconds...\n", d.Body, waitFor)

			time.Sleep(time.Second * time.Duration(waitFor))
			err = d.Ack(false)
			checkErr(err)
		}
	}()

	fmt.Printf("[*] Waiting for messages. To exit precc CTRL+C")
	<-make(chan bool)

}

func producer() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	checkErr(err)
	defer conn.Close()

	// open a channel
	ch, err := conn.Channel()
	checkErr(err)
	defer ch.Close()

	// open a queue
	queue, err := ch.QueueDeclare(
		"racing",
		true,
		false,
		false,
		false,
		nil,
	)
	checkErr(err)

	fmt.Println("Running producer, precc CTRL+C to stop")

	go func() {
		for i := 0; true; i++ {
			text := fmt.Sprintf("Message to process: %v", i)
			msg := amqp.Publishing{
				ContentType: "/text/plain",
				Body:        []byte(text),
			}
			err = ch.Publish(
				"", // default exchange
				queue.Name,
				false,
				false,
				msg,
			)
			checkErr(err)
			fmt.Printf("Published \"%s\"\n", text)

			waitFor := rand.Intn(4) + 1
			time.Sleep(time.Duration(waitFor) * time.Second)
		}
	}()
	<-make(chan bool)
}

var modeMap map[string]func() = map[string]func(){
	"consumer": consumer,
	"producer": producer,
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Invalid amount of command line arguments")
		os.Exit(1)
	}

	mode := os.Args[1]
	fun, ok := modeMap[mode]
	if !ok {
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}

	fun()
}
