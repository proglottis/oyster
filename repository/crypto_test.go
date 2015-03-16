package repository

import (
	"testing"
)

func TestGpgRepoSecureKeyRing(t *testing.T) {
	repo := NewGpgRepo("gpghome")
	el, err := repo.SecureKeyRing([]string{"test@example.com"})
	if err != nil {
		t.Error(err)
	}
	if len(el) != 1 {
		t.Error("expected 1 entity, got", len(el))
	}
}

func TestGpgRepoPublicKeyRing(t *testing.T) {
	repo := NewGpgRepo("gpghome")
	el, err := repo.PublicKeyRing([]string{"test@example.com"})
	if err != nil {
		t.Error(err)
	}
	if len(el) != 1 {
		t.Error("expected 1 entity, got", len(el))
	}
}

func TestEntityMatchesId(t *testing.T) {
	el, err := ReadKeyRing("gpghome/pubring.gpg")
	if err != nil {
		t.Error(err)
	}
	entity := el[0]
	keyid := entity.PrimaryKey.KeyIdString()
	keyidshort := entity.PrimaryKey.KeyIdShortString()

	if !EntityMatchesId(entity, "test@example.com") {
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
