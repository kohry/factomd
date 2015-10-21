package primitives_test

import (
	"testing"

	. "github.com/FactomProject/factomd/common/primitives"
)

func TestSignatureMisc(t *testing.T) {
	priv1 := new(PrivateKey)

	err := priv1.GenerateKey()
	if err != nil {
		t.Fatalf("%v", err)
	}

	msg1 := "Test Message Sign1"
	msg2 := "Test Message Sign2"

	sig1 := priv1.Sign([]byte(msg1))

	if !sig1.Verify([]byte(msg1)) {
		t.Fatalf("sig1.Verify retuned false")
	}

	if sig1.Verify([]byte(msg2)) {
		t.Fatalf("sig1.Verify retuned true")
	}

	sigBytes := append(sig1.Pub.Key[:], sig1.Sig[:]...)

	sig2 := UnmarshalBinarySignature(sigBytes)

	if !sig2.Verify([]byte(msg1)) {
		t.Fatalf("sig2.Verify retuned false")
	}

	if sig2.Verify([]byte(msg2)) {
		t.Fatalf("sig2.Verify retuned true")
	}

	pub := sig2.Pub.Key[:]
	pub2 := sig2.Key()

	if len(pub) != len(pub2) {
		t.Error("Public key length mismatch")
	}
	for i := range pub {
		if pub[i] != pub2[i] {
			t.Error("Pub keys are not identical")
		}
	}
}
