package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func generateRandomKey(size int) string {
	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("Failed to generate random key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

func main() {
	secretKey := generateRandomKey(32)
	log.Println("Generated secret key:", secretKey)
}
