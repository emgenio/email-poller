package main

import (
  "emailPoller"
)

func main() {
  emailPoller := emailPoller.Create("../config.yaml")

  emailPoller.Initialize()
  emailPoller.Start()
  if false {
    _ = emailPoller
  }
}