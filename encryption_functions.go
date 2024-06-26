package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	random "math/rand/v2"
	"os"
	"strings"
	"log"
	//"fmt"
)

// Encrypt the specified data with a specified key
func encrpytData(key []byte, data []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(aes)
    if err != nil {
        return nil, err
    }

    // We need a 12-byte nonce for GCM (modifiable if you use cipher.NewGCMWithNonceSize())
    // A nonce should always be randomly generated for every encryption.
    nonce := make([]byte, gcm.NonceSize())
    _, err = rand.Read(nonce)
    if err != nil {
        return nil, err
    }

    // ciphertext here is actually nonce+ciphertext
    // So that when we decrypt, just knowing the nonce size
    // is enough to separate it from the ciphertext.
    ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

    return ciphertext, nil
}

// Decrypt the specified data with a specified key
func decryptData(key []byte, data []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(aes)
    if err != nil {
        return nil, err
    }

    // Since we know the ciphertext is actually nonce+ciphertext
    // And len(nonce) == NonceSize(). We can separate the two.
    nonceSize := gcm.NonceSize()
    nonce, data := data[:nonceSize], data[nonceSize:]

    plaintext, err := gcm.Open(nil, []byte(nonce), []byte(data), nil)
    if err != nil {
        return nil, err
    }

    return plaintext, nil
}

// Either get an existing key or generate a new one
func generateEncryptionKey(keyFilePath string) (error) {
	content, err := os.ReadFile(keyFilePath)
	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file") {
			log.Println("System cannot find an existing key - creating a new one.")
		} else {
			return err
		}
	}

	if len(content) < 32 {
		var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

		// Handle opening / creating the new key file
		file, err := os.OpenFile("keys/main.dat", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}

		// Generate a new byte key at random
		b := make([]byte, 32)
		for i := range b {
			b[i] = letters[random.IntN(len(letters))]
		}
		newKey := string(b)

		// Set the key to environment
		setEnvErr := os.Setenv("EK", newKey)
		if setEnvErr != nil {
			return setEnvErr
		}

		// Handle errors with writing the key to file
		_, fileWriteErr := file.Write(b)
		if fileWriteErr != nil {
			return fileWriteErr
		}

		defer file.Close()
	} else {
		// Set the key to environment
		setEnvErr := os.Setenv("EK", string(content))
		if setEnvErr != nil {
			return setEnvErr
		}
	}

	return nil
}

// Function to rotate the encryption key, should be used periodicallys
func rotateEncryptionKey(keyFilePath string) (error) {return nil}

// Function to generate private key
func generatePrivateKey() ([]byte, error) {
	// generate a private key
	privateKey, privateKeyErr := rsa.GenerateKey(rand.Reader, 2048)
    if privateKeyErr != nil {
        return nil, privateKeyErr
    }

	// Convert the private key to a byte array, and return it
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	return privateKeyBytes, nil
}

// Function to generate a public key, based on a private key
func generatePublicKey(privateKeyBytes []byte) ([]byte, error) {
	// Convert the private key from its bytes to a working private key
	privateKey, privateKeyErr := x509.ParsePKCS1PrivateKey(privateKeyBytes)
    if privateKeyErr != nil {
        return nil, privateKeyErr
    }

	// Generate the public key from the private key, and convert it into a byte array
	publicKeyBytes, publicKeyErr := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
    if publicKeyErr != nil {
        return nil, publicKeyErr
    }

	return publicKeyBytes, nil
}

// Encrypts data with a public key
func encryptWithPublicKey(publicKeyBytes []byte, dataToEncrypt []byte) ([]byte, error) {
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
    if err != nil {
        return nil, err
    }

    encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey.(*rsa.PublicKey), dataToEncrypt)
    if err != nil {
        return nil, err
    }

	return encryptedData, nil
}

// Decrypts data with a private key
func decryptWithPrivateKey(privateKeyBytes []byte, encryptedData []byte) ([]byte, error) {
	// Convert the private key bytes into a working private key
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
    if err != nil {
        return nil, err
    }

	// Decrypt the data and return it
    decryptedData, decryptErr := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
    if decryptErr != nil {
        return nil, decryptErr
    }

	return decryptedData, nil
}

// Use this to confirm if the public key provided matches the public key generated by the private key
func confirmPublicKey(publicKeyBytes []byte, privateKeyBytes []byte) (bool, error) {
	publicKey, publicKeyErr := x509.ParsePKIXPublicKey(publicKeyBytes)
    if publicKeyErr != nil {
        return false, publicKeyErr
    }

	privateKey, privateKeyErr := x509.ParsePKCS1PrivateKey(privateKeyBytes)
    if privateKeyErr != nil {
        return false, privateKeyErr
    }

	return privateKey.PublicKey.Equal(publicKey.(*rsa.PublicKey)), nil
}