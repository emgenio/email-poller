package main

import (
  "imapWrapper"
)

func main() {
  imapClient := imapWrapper.Create("../config.yaml")

  imapClient.Connect()
  if false {
    _ = imapClient
  }
}