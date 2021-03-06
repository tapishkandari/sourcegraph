package secret

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"os"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
)

// gatherKeys splits the comma-separated encryption data into its potential two components:
// primary and secondary keys, where the first key is assumed to be the primary key.
func gatherKeys(data []byte) (primaryKey, secondaryKey []byte, err error) {
	parts := bytes.Split(data, []byte(","))
	if len(parts) > 2 {
		return nil, nil, errors.Errorf("expect at most two encryption keys but got %d", len(parts))
	}
	if len(parts) == 1 {
		return parts[0], nil, nil
	}
	return parts[0], parts[1], nil
}

var initErr error
var initOnce sync.Once

// Init creates the defaultEncryptor by ingesting user encryption key(s).
// For production deployments, the secret value CAN ONLY be generated by the user and loaded via a file or env var.
// For single server docker deployments, we will generate the secret file and write it to disk.
func Init() error {
	initOnce.Do(func() {
		initErr = initDefaultEncryptor()
	})
	return initErr
}

// defaultEncryptor is configured during init, if no keys are provided it will implement noOpEncryptor.
var defaultEncryptor encryptor = noOpEncryptor{}

// NOTE: MockDefaultEncryptor should only be called in tests where a random encryptor is
// needed to test transparent encryption and decryption.
func MockDefaultEncryptor() {
	defaultEncryptor = newAESGCMEncodedEncryptor(mustGenerateRandomAESKey(), nil)
}

const sourcegraphsSecretFile = "SOURCEGRAPH_SECRET_FILE"

func initDefaultEncryptor() error {
	// Set the default location if none exists
	secretFile := os.Getenv(sourcegraphsSecretFile)
	if secretFile == "" {
		secretFile = "/etc/sourcegraph/token"
	}

	fileInfo, err := os.Stat(secretFile)
	if err != nil {
		defaultEncryptor = noOpEncryptor{}
		log15.Warn("No encryption initialized")
		return nil
	}

	perm := fileInfo.Mode().Perm()
	if perm != os.FileMode(0400) {
		return errors.New("key file permissions are not 0400")
	}

	encryptionKey, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return errors.Wrapf(err, "couldn't read file %s", sourcegraphsSecretFile)
	}
	if len(encryptionKey) < requiredKeyLength {
		return errors.Errorf("key length of %d characters is required", requiredKeyLength)
	}

	primaryKey, secondaryKey, err := gatherKeys(encryptionKey)
	if err != nil {
		return errors.Wrap(err, "gather keys")
	}

	defaultEncryptor = newAESGCMEncodedEncryptor(primaryKey, secondaryKey)
	log15.Info("Database secrets encryption initialized")
	return nil
}

// generateRandomAESKey generates a random key that can be used for AES-256 encryption.
func generateRandomAESKey() ([]byte, error) {
	b := make([]byte, requiredKeyLength)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// mustGenerateRandomAESKey generates a random AES key and panics for any error.
func mustGenerateRandomAESKey() []byte {
	key, err := generateRandomAESKey()
	if err != nil {
		panic(err)
	}
	return key
}
