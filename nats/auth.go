package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

// AuthFromSeed takes a seed and returns a nats options that is an Nkey
// authenticator.
func AuthFromSeed(seed string) (nats.Option, error) {
	kp, err := nkeys.FromSeed([]byte(seed))
	if err != nil {
		return nil, err
	}

	pub, err := kp.PublicKey()
	if err != nil {
		return nil, fmt.Errorf("unable to extract public key from seed: %s", err.Error())
	}

	if !nkeys.IsValidPublicUserKey(pub) {
		return nil, fmt.Errorf("nats: Not a valid nkey user seed")
	}

	sigCB := func(nonce []byte) ([]byte, error) {
		kp, err := nkeys.FromSeed([]byte(seed))
		if err != nil {
			return nil, err
		}
		// Wipe our key on exit.
		defer kp.Wipe()

		sig, _ := kp.Sign(nonce)
		return sig, nil
	}

	return nats.Nkey(pub, sigCB), nil
}
