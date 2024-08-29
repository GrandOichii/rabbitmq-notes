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

func consumer(args []string) {
	if len(args) != 1 {
		fmt.Println("Invalid amount of command line arguments")
		os.Exit(1)
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	checkErr(err)
	defer conn.Close()

	// open a channel
	ch, err := conn.Channel()
	checkErr(err)
	defer ch.Close()

	// open an exchange
	exName := "pubsub"
	err = ch.ExchangeDeclare(
		exName,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	checkErr(err)

	// open a queue
	queue, err := ch.QueueDeclare(
		fmt.Sprintf("pubsub-consumer-%s", args[0]),
		true,
		false,
		false,
		false,
		nil,
	)
	checkErr(err)

	// bind the queue
	ch.QueueBind(
		queue.Name,
		"",
		exName,
		false,
		nil,
	)

	// receive messages
	msgs, err := ch.Consume(
		queue.Name,
		"",
		// fmt.Sprintf("consumer%v", args[0]),
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	checkErr(err)

	go func() {
		for d := range msgs {
			fmt.Printf("Received: \"%s\"\n", d.Body)
		}
	}()

	fmt.Printf("[*] Waiting for messages. To exit precc CTRL+C")
	<-make(chan bool)
}

func producer(args []string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	checkErr(err)
	defer conn.Close()

	// open a channel
	ch, err := conn.Channel()
	checkErr(err)
	defer ch.Close()

	// open an exchange
	exName := "pubsub"
	err = ch.ExchangeDeclare(
		exName,
		"fanout",
		false,
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
				exName, // default exchange
				"",
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

var modeMap map[string]func([]string) = map[string]func([]string){
	"consumer": consumer,
	"producer": producer,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Invalid amount of command line arguments")
		os.Exit(1)
	}

	mode := os.Args[1]
	fun, ok := modeMap[mode]
	if !ok {
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}

	fun(os.Args[2:])
}
