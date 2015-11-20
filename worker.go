package main

import (
  "fmt"
  "github.com/streadway/amqp"
  "log"
  "encoding/json" 
)

func failOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}

func main() {
  conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
  failOnError(err, "Failed to connect to RabbitMQ")
  defer conn.Close()

  ch, err := conn.Channel()
  failOnError(err, "Failed to open a channel")
  defer ch.Close()

  q, err := ch.QueueDeclare(
    "email_poller_messages", // name
    true,         // durable
    false,        // delete when unused
    false,        // exclusive
    false,        // no-wait
    nil,          // arguments
  )
  failOnError(err, "Failed to declare a queue")

  err = ch.Qos(
    1,     // prefetch count
    0,     // prefetch size
    false, // global
  )
  failOnError(err, "Failed to set QoS")

  msgs, err := ch.Consume(
    q.Name, // queue
    "",     // consumer
    false,  // auto-ack
    false,  // exclusive
    false,  // no-local
    false,  // no-wait
    nil,    // args
  )
  failOnError(err, "Failed to register a consumer")

  forever := make(chan bool)

  go func() {
    for d := range msgs {
      log.Println("Received a message:")
      msg := decodeImapMessage(d.Body)
      msg.Dump()
      d.Ack(false)
    }
  }()

  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
  <-forever
}

func decodeImapMessage(message []byte) (ImapMessage) {
  msg := ImapMessage{}
  err := json.Unmarshal(message, &msg)
  if err != nil {
    panic (err)
  }
  return msg
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