package main

import "booking"

func main() {
	client := &booking.Client{}
	client.LoginAll()
	client.SetFormInfo()
	client.Submit()

	select {}
}
