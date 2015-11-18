package main

import (
  "imapWrapper"
)

func main() {
  imapClient := imapWrapper.Create("../config.yaml")

  imapClient.Connect()
  imapClient.Login()
  imapClient.SelectBox("INBOX", false) // true for Read only
  imapClient.FetchAllMessages()
  if false {
    _ = imapClient
  }
}