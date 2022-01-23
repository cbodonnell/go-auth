package hash

import "testing"

func TestBCryptHash(t *testing.T) {
	h := NewBCryptHash(14)
	testPassword := "password"
	testHash, err := h.Generate(testPassword)
	if err != nil {
		t.Fatal("failed to generate hash", err)
	}
	err = h.Check(testHash, testPassword)
	if err != nil {
		t.Fatal("failed to validate hash", err)
	}
}
