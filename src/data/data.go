package data

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	// Postgres driver.
	_ "github.com/lib/pq"
)

// DB is used in data package to perform database related operations.
var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		"postgres", "postgresdb", "db1"))

	if err != nil {
		log.Fatal(err)
	}
}

// createUUID function creates a unique identifier
// that can be used as session id.
func createUUID() (uuid string) {
	// create a byte slice of length 32.
	b := make([]byte, 32)

	// use rand.Reader
	rdr := rand.Reader

	// read random data into b.
	_, err := rdr.Read(b)

	// error handling.
	if err != nil {
		panic(err)
	}

	// this can be used as a session uuid.
	uuid = base64.URLEncoding.EncodeToString(b)

	return
}

// Encrypt function encrypts plain text and returns base64 string.
func Encrypt(plainTextInBytes []byte, keyInBytes []byte) (cipherString string, err error) {

	// 1. Create new cipher using aes. Pass key
	c, err := aes.NewCipher(keyInBytes)

	panicErrs(err)

	// 2. Create new gcm using cipher. Pass cipher created in step 1.
	gcm, err := cipher.NewGCM(c)

	panicErrs(err)

	// 3: Create a nonce of size from gcm created in step 2.
	nonce := make([]byte, gcm.NonceSize())

	// 4: Read random data into nonce byte array.
	_, err = io.ReadFull(rand.Reader, nonce)

	panicErrs(err)

	// 5. gcm's seal method returns encrypted data in bytes.
	cipherInBytes := gcm.Seal(nonce, nonce, plainTextInBytes, nil)

	cipherString = base64.StdEncoding.EncodeToString(cipherInBytes)

	return
}

// Decrypt function decrypts base64 string into plaintext.
func Decrypt(cipherString string, keyInBytes []byte) (decData []byte, err error) {

	// 0. Decode base64 cipher string into bytes.
	cipherTextInBytes, err := base64.StdEncoding.DecodeString(cipherString)

	panicErrs(err)

	// 1. Create a new cipher using aes and key.
	c, err := aes.NewCipher(keyInBytes)

	// 2. Create new GCM using c created in step 1.
	gcm, err := cipher.NewGCM(c)

	panicErrs(err)

	// 3. Get Nonce size from GCM.
	nonceSize := gcm.NonceSize()

	if len(cipherTextInBytes) < nonceSize {
		panicErrs(errors.New("ciphertext is too short"))
	}

	nonce, cipherTextInBytes := cipherTextInBytes[:nonceSize], cipherTextInBytes[nonceSize:]

	decData, err = gcm.Open(nil, nonce, cipherTextInBytes, nil)

	return
}

func panicErrs(err error) {
	if err != nil {
		panic(err)
	}
}
