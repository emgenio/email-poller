package main

import (
  "fmt"
  // "time"
  "log"
  "encoding/json"
  "github.com/streadway/amqp"
  "./imap"
  // "time"
  // "github.com/mxk/go-imap/imap"
)

var (
  client  *imapClient.GoImapClient
  err     error
)

const (
  hostname = "localhost"
  user     = "axl"
  password = "fuckmalife"
)

func encodeImapMessage(message imapClient.GoImapMessage) ([]byte) {
  out, err := json.Marshal(message)
  if err != nil {
    panic (err)
  }
  return out
}

func main() {
  // instanciate new GoImapClient
  client = imapClient.NewClient(hostname, 143, user, password)
  // Connect to server (Dial)
  err = client.Connect()
  // defer client.Logout(30 * time.Second)
  client.Client.Data = nil

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
  go func() {
    conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
    failOnError(err, "Failed to connect to RabbitMQ")
    defer conn.Close()

    ch, err := conn.Channel()
    failOnError(err, "Failed to open a channel")
    defer ch.Close()

    q, err := ch.QueueDeclare(
      "email_poller_messages", // name
      true,    // durable
      false,   // delete when usused
      false,   // exclusive
      false,   // no-wait
      nil,     // arguments
    )
    failOnError(err, "Failed to declare a queue")

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
            ContentType:  "text/plain",
            Body:         encodeImapMessage(msg),
          })
        failOnError(err, "Failed to publish a message")
        count += 1
      }
      fmt.Println("Messages pushed to RabbitMQ:", count)
    }
  }()

  // c := client.Client
  for {
    _, err := client.WaitForNotifications()

    // fmt.Println("Setting Client in Idle state...")
    // _, err = c.Idle()
    // if err != nil {
    //   fmt.Println("Error when Idling:", err)
    // }
    // fmt.Println("Waiting for notifications (30 sec timeout not to disconnect the c, RFC says)")
    // c.Recv(2 * time.Second)
    // if err != nil {
    //   fmt.Println("Error when Recv:", err)
    // }
    // fmt.Println("Notifications received...")
    // fmt.Println("Terminating Idle state...")
    // _, err = imap.Wait(c.IdleTerm())
    // if err != nil {
    //   fmt.Println("Error when Terminating Idle state:", err)
    // }

    fmt.Println("rcv message")
    // ids := []uint32{}
    // ids = client.RetrieveMessageIds()
    client.Client.Data = nil
    if false {
      // _=ids
      _=err
    }
    // messages, error := client.RetrieveMessagesFromIds(ids)
    // for _, msg := range messages {
    //   msg.Dump()
    // }
    // if error != nil {
    //   fmt.Println("Error FetchMessagesFromIds:", err)
    // }
    // if len(messages) > 0 {
    //   messagesChan <- messages
    //   client.ExpungeMessages(messages)
    // }
  }
}

func failOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}