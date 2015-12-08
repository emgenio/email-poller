package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/catuss-a/imap"
	"github.com/streadway/amqp"
)

var (
	err    error
	config PollerConfig
)

const (
	Timeout           = 30 * time.Second
	defaultConfigFile = "./config.yaml"
)

func fatalOnError(err error, msg string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func pushIncomingMessagesToQueue(messagesChan chan []imapClient.GoImapMessage) {
	conn, err := amqp.Dial(config.Amqp.Hostname)
	fatalOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	fatalOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		config.Amqp.MessageQueue, // queue
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	fatalOnError(err, "Failed to declare a queue")

	for newMessages := range messagesChan {
		count := 0
		for _, msg := range newMessages {
			err = ch.Publish(
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,
				amqp.Publishing{
					DeliveryMode: amqp.Persistent,
					ContentType:  "text/plain",
					Body:         msg.EncodeAsBytes(),
				})

			if err != nil {
				fatalOnError(err, "Failed to publish a message")
			} else {
				count += 1
			}
		}
		fmt.Println("Amount of Messages pushed to RabbitMQ:", count)
	}
}

func monitorMailbox(client *imapClient.GoImapClient,
	messagesChan chan []imapClient.GoImapMessage) {
	for {
		client.WaitForNotifications(Timeout)
		ids := []uint32{}
		ids = client.RetrieveMessageIds()
		client.Client.Data = nil
		messages := client.RetrieveMessagesFromIds(ids)
		if len(messages) > 0 {
			messagesChan <- messages
			for _, msg := range messages {
				msg.Dump()
			}
			client.ExpungeMessages(messages)
		}
	}
}

type PollerConfig struct {
	Imap struct {
		Hostname string
		Port     uint32
		Login    string
		Password string
	}
	Amqp struct {
		Hostname     string
		MessageQueue string
	}
}

func init() {
	configPath := flag.String("config", defaultConfigFile, "path to the configuration file.")
	flag.Parse()
	fmt.Println("Loading configuration file", *configPath)
	loadConfig(*configPath, &config)
}

func main() {
	// instanciate new GoImapClient
	client := imapClient.NewClient(config.Imap.Hostname, config.Imap.Port, config.Imap.Login, config.Imap.Password)

	// Connect to server (Dial)
	err = client.Connect()
	defer client.Logout(30 * time.Second)

	// Authenticate
	_, err = client.Login()

	// Open a mailbox (synchronous command - no need for imap.Wait)
	client.SelectMailBox("INBOX", false)

	// Print box status
	fmt.Print("\nMailbox status:\n", client.Client.Mailbox)

	if !client.SupportIdleCap() {
		fmt.Println("Error: Server does not support IDLE state")
		return
	}

	messagesChan := make(chan []imapClient.GoImapMessage)

	// Wait for Incoming ImapMessages and send them to RabbitMQ
	go pushIncomingMessagesToQueue(messagesChan)

	// Monitor mailbox and push incoming messages to Channel
	monitorMailbox(client, messagesChan)
}
