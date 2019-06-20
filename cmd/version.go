package cmd

import (
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run:   runVersion,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func send() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("error %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("error %v", err)
	}
	defer func() {
		_ = ch.Close()
	}()
	q, err := ch.QueueDeclare(
		"whale-fs", // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatalf("error %v", err)
	}
	body := "Hello World!"
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		log.Fatalf("error %v", err)
	}
}

func receive() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("error %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("error %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"whale-fs", // name
		true,       // durable
		false,      // delete when usused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)

	if err != nil {
		log.Fatalf("error %v", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Fatalf("error %v", err)
	}
	for d := range msgs {
		log.Printf("Received a message: %s", d.Body)
		err = ch.Ack(d.DeliveryTag, false)
		if err != nil {
			log.Warnf("error %v", err)
		}
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	// TODO(benjamin): add release version
	fmt.Printf("0.0.1")
	send()

	receive()
}
