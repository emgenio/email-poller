package main

import (
  "../imap"
  "time"
)

func main() {
  c := imap.NewClient("localhost", 143, "axl", "fuckmalife")
  _, err := c.Connect()
  defer c.Logout(1 * time.Second)

  if false {
    _ = c
    _ = err
  }
}