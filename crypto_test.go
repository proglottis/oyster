package main

import (
	"testing"

	"code.google.com/p/go.crypto/openpgp"
)

func TestEntityMatchesId(t *testing.T) {
	entity, _ := openpgp.NewEntity("Bob", "", "bob@example.com", nil)
	entity.PrivateKey = nil
	keyid := entity.PrimaryKey.KeyIdString()
	keyidshort := entity.PrimaryKey.KeyIdShortString()

	if !EntityMatchesId(entity, "bob@example.com") {
		t.Error("must match by email")
	}
	if !EntityMatchesId(entity, keyid) {
		t.Error("must match by key ID")
	}
	if !EntityMatchesId(entity, keyidshort) {
		t.Error("must match by short key ID")
	}
	if EntityMatchesId(entity, "no_match") {
		t.Error("should not match")
	}
}
