package main

import (
	"fmt"
	"os"

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
	exName := "routing"
	err = ch.ExchangeDeclare(
		exName,
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)
	checkErr(err)

	route := args[0]

	// open a queue
	queue, err := ch.QueueDeclare(
		"",
		true,
		true, // will delete itself after closing
		false,
		false,
		nil,
	)
	checkErr(err)

	// bind the queue
	ch.QueueBind(
		queue.Name,
		route,
		exName,
		false,
		nil,
	)

	// receive messages
	msgs, err := ch.Consume(
		queue.Name,
		"",    // will auto-generate
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

func publishTo(args []string) {
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
	exName := "routing"
	err = ch.ExchangeDeclare(
		exName,
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)
	checkErr(err)

	route := args[0]

	text := fmt.Sprintf("Sending message to route %s", route)
	msg := amqp.Publishing{
		ContentType: "/text/plain",
		Body:        []byte(text),
	}
	err = ch.Publish(
		exName,
		route,
		false,
		false,
		msg,
	)
	checkErr(err)
}

var modeMap map[string]func([]string) = map[string]func([]string){
	"consumer":  consumer,
	"publishto": publishTo,
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
