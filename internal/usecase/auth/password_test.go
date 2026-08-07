package auth

import "testing"

const demoPasswordHash = "$argon2id$v=19$m=19456,t=2,p=1$H7eaNLiQRPnkW97cUoyUBw$1gSVVGrLCuY1ORViVB7c8CgI29gueEN7WkKL+4dsm2E"

func TestArgon2Hasher(t *testing.T) {
	hasher := argon2Hasher{}
	hash, err := hasher.Hash("надёжный-пароль")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	valid, err := hasher.Compare("надёжный-пароль", hash)
	if err != nil || !valid {
		t.Fatalf("compare correct password: valid=%v err=%v", valid, err)
	}
	valid, err = hasher.Compare("другой-пароль", hash)
	if err != nil {
		t.Fatalf("compare incorrect password: %v", err)
	}
	if valid {
		t.Fatal("incorrect password accepted")
	}
}

func TestDemoPasswordHash(t *testing.T) {
	valid, err := (argon2Hasher{}).Compare("avito2026", demoPasswordHash)
	if err != nil || !valid {
		t.Fatalf("demo password hash is invalid: valid=%v err=%v", valid, err)
	}
}
