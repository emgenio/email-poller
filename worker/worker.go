package main

import (
  "fmt"
  "log"

  gconf "../config"
  "github.com/streadway/amqp"
  "github.com/keighl/mandrill"
  "github.com/emgenio/email-poller/imap"
)

func fatalOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}

type WorkerConfig struct {
  Amqp struct {
    Hostname      string
    MessageQueue  string
  }
  Mandrill struct {
    ClientKey   string
    From        string
  }
}

var (
  cfg   WorkerConfig
)

func init() {
  gconf.LoadConfig("./worker.yaml", &cfg)
}

func main() {
  conn, err := amqp.Dial(cfg.Amqp.Hostname)
  fatalOnError(err, "Failed to connect to RabbitMQ")
  defer conn.Close()

  ch, err := conn.Channel()
  fatalOnError(err, "Failed to open a channel")
  defer ch.Close()

  q, err := ch.QueueDeclare(
    cfg.Amqp.MessageQueue, // name
    true,         // durable
    false,        // delete when unused
    false,        // exclusive
    false,        // no-wait
    nil,          // arguments
  )
  fatalOnError(err, "Failed to declare a queue")

  err = ch.Qos(
    1,     // prefetch count
    0,     // prefetch size
    false, // global
  )
  fatalOnError(err, "Failed to set QoS")

  msgs, err := ch.Consume(
    q.Name, // queue
    "",     // consumer
    false,  // auto-ack
    false,  // exclusive
    false,  // no-local
    false,  // no-wait
    nil,    // args
  )
  fatalOnError(err, "Failed to register a consumer")

  forever := make(chan bool)

  go func() {
    mclient := mandrill.ClientWithKey(cfg.Mandrill.ClientKey)
    for rawMsg := range msgs {
      decodedMsg := imapClient.NewMessageFromBytes(rawMsg.Body)
      forwardMessage(mclient, decodedMsg)
      rawMsg.Ack(false)
    }
  }()
  <- forever
}

func forwardMessage(mclient *mandrill.Client, message *imapClient.GoImapMessage) {
  msg := &mandrill.Message{}
  msg.AddRecipient("axel.catusse@gmail.com", "Axel Catusse", "to")

  msg.FromEmail = cfg.Mandrill.From
  msg.Subject   = message.Subject
  msg.Text      = message.Body

  _, err := mclient.MessagesSend(msg)
  if err != nil {
    fmt.Println("forwardMessage:", err)
  } else {
    fmt.Println("Message forwarded successfully")
  }
}
