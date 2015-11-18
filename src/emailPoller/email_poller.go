package emailPoller

import (
  "imapWrapper"
  "time"
)

const (
  TimeOut = time.Minute * 30
)

type emailPoller struct {
  ImapClient *ImapWrapper
  MessagesToQ chan[]IWMessage

  error error
}

func (obj *emailPoller) Initialize() {
  obj.ImapClient.Connect()
  obj.ImapClient.Login()
  obj.ImapClient.SelectBox(ImapClient.Config.Mbox, false)
}

func (obj *emailPoller) PushMessagesToQueue() {
  for {
    messages <- obj.MessagesToQ

    for _, message := range messages {
      // push to RABBIT
    }
  }
}

func (obj *emailPoller) Start() {
  go obj.PushMessagesToQueue()

  for {
    // wait for incoming messages
    // fill channel
    // Tag them as Seen/Deleted
  }
}

func Stop() {}
func init {
}

func Create(cfg string) {
  return &emailPoller{
    ImapClient: imapWrapper.Create(cfg),
    MessagesToQ: make(chan []IWMessage)
  }
}