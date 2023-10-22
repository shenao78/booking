package main

import "booking"

func main() {
	client := &booking.Client{}
	client.Login()
	client.SetFormInfo()

	select {}
}
