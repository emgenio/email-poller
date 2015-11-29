package main

import (
  "fmt"
  "time"
  "log"
  gconf "../config"
  "github.com/streadway/amqp"
  "github.com/emgenio/email-poller/imap"
)

var (
  err     error
  cfg     PollerConfig
)

const (
  Timeout = 30 * time.Second
)

func fatalOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}

func pushIncomingMessagesToQueue(messagesChan chan []imapClient.GoImapMessage) {
  conn, err := amqp.Dial(cfg.Amqp.Hostname)
    fatalOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    fatalOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
      "", // name
      true,    // durable
      false,   // delete when usused
      false,   // exclusive
      false,   // no-wait
      nil,     // arguments
    )
    fatalOnError(err, "Failed to declare a queue")

    for newMessages := range messagesChan {
      count := 0
      for _, msg := range newMessages {
        err = ch.Publish(
          "",           // exchange
          q.Name,       // routing key
          false,        // mandatory
          false,
          amqp.Publishing{
            DeliveryMode: amqp.Persistent,
            ContentType:  "toext/plain",
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
    Hostname    string
    Port        uint32
    Login       string
    Password    string
  }
  Amqp struct {
    Hostname      string
    MessageQueue  string
  }
}

func init() {
  gconf.LoadConfig("./email-poller.yaml", &cfg)
}

func main() {
  // instanciate new GoImapClient
  client := imapClient.NewClient(cfg.Imap.Hostname, cfg.Imap.Port, cfg.Imap.Login, cfg.Imap.Password)

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
