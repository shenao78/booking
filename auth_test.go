package booking

import "testing"

func TestWriteAuth(t *testing.T) {
	user := "33082519910910454X"
	pass := "Ww2018$$$"
	writeUserPass(user, pass)
	readUser, readPass, err := readUserPass()
	if err != nil {
		t.Fatal(err)
	}
	if readUser != user && readPass != pass {
		t.Error("failed to read auth file")
	}
}
