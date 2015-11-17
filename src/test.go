package main

import (
  "fmt"
  "imapWrapper"
)

func main() {
  imapClient := imapWrapper.Create("../config.yaml")

  fmt.Printf("test")

  if false {
    _ = imapClient
  }
}