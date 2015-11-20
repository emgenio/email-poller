package main

import (
  "fmt"
  "time"
  "log"
  "encoding/json"
  "github.com/mxk/go-imap/imap"
  "github.com/streadway/amqp"
)

var (
  client  *imap.Client
  cmd     *imap.Command
  rsp     *imap.Response
  err     error
)

const (
  hostname = "localhost"
  user     = "axl"
  password = "fuckmalife"
)

func encodeImapMessage(message ImapMessage) ([]byte) {
  out, err := json.Marshal(message)
  if err != nil {
    panic (err)
  }
  return out
}

func main() {
  client, err = imap.Dial(hostname)
  defer client.Logout(30 * time.Second)

  // Print server greeting (first response in the unilateral server data queue)
  fmt.Println("Server says:", client.Data[0].Info)
  client.Data = nil

  fmt.Println("Client State:", client.State())
  // Authenticate
  if client.State() == imap.Login {
    cmd, err = imap.Wait(client.Login(user, password))
  }

  // Open a mailbox (synchronous command - no need for imap.Wait)
  client.Select("INBOX", true)
  fmt.Print("\nMailbox status:\n", client.Mailbox)

  if !client.Caps["IDLE"] {
    fmt.Println("Error: Server does not support IDLE state")
    return
  }

  messagesChan := make(chan []ImapMessage)

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
        msg.Dump()
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

  for {
    fmt.Println("Setting Client in Idle state...")
    cmd, err = client.Idle()
    if err != nil {
      fmt.Println("Error when Idling:", err)
    }
    fmt.Println("Waiting for notifications (30 sec timeout not to disconnect the Client, RFC says)")
    client.Recv(30 * time.Second)
    if err != nil {
      fmt.Println("Error when Recv:", err)
    }
    fmt.Println("Notifications received...")
    fmt.Println("Terminating Idle state...")
    cmd, err = imap.Wait(client.IdleTerm())
    if err != nil {
      fmt.Println("Error when Terminating Idle state:", err)
    }

    ids := []uint32{}
    for _, resp := range client.Data {
      switch resp.Label {
      case "EXISTS":
        ids = append(ids, imap.AsNumber(resp.Fields[0]))
      }
    }
    fmt.Println("Messages IDS received:", ids)
    client.Data = nil

    messages, error := FetchMessagesFromIds(client, ids)
    if error != nil {
      fmt.Println("Error FetchMessagesFromIds:", err)
    }
    if len(messages) > 0 {
      messagesChan <- messages
      ExpungeMessages(client, messages)
    }
  }


  if false {
    _ = err
    _ = cmd
  }
}

type ImapMessage struct {
  UID     uint32
  Header  []byte
  Body    []byte
}

func (obj *ImapMessage) Dump() {
  fmt.Println("-------------------------------------------------")
  fmt.Println("UID:\n", obj.UID)
  fmt.Println("HEADER:\n", string(obj.Header))
  fmt.Println("BODY:\n", string(obj.Body))
  fmt.Println("-------------------------------------------------")
}

func FetchMessagesFromIds(c *imap.Client, ids []uint32) ([]ImapMessage, error) {
  messages := []ImapMessage{}

  if len(ids) > 0 {
    set, _ := imap.NewSeqSet("")
    set.AddNum(ids...)

    cmd, err := imap.Wait(c.Fetch(set, "UID", "RFC822.HEADER", "RFC822.TEXT"))
    if err != nil {
      return messages, fmt.Errorf("An error ocurred while fetching unread messages data. ", err)
    }

    for _, msg := range cmd.Data {
      attrs := msg.MessageInfo().Attrs
      message := ImapMessage{
        UID:    imap.AsNumber(attrs["UID"]),
        Header: imap.AsBytes(attrs["RFC822.HEADER"]),
        Body:   imap.AsBytes(attrs["RFC822.TEXT"]),
      }
      messages = append(messages, message)
    }
  }

  return messages, nil
}

func ExpungeMessages(c *imap.Client, messages []ImapMessage) (error) {
  set, _ := imap.NewSeqSet("")

  for _, message := range messages {
    set.AddNum(message.UID)
  }
  _, err := imap.Wait(c.UIDStore(set, "+FLAGS", imap.NewFlagSet(`\Deleted`)))
  if err != nil {
    fmt.Println("Error UIDStore:", err)
  }
  _, err = imap.Wait(c.Expunge(nil))
  if err != nil {
    fmt.Println("Error Expunge:", err)
  }
  return err
}

func failOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}