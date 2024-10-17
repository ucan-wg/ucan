package did

import (
	"fmt"
	"strings"

	crypto "github.com/libp2p/go-libp2p/core/crypto"
	mbase "github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"
	varint "github.com/multiformats/go-varint"
)

// Signature algorithms from the [did:key specification]
//
// [did:key specification]: https://w3c-ccg.github.io/did-method-key/#signature-method-creation-algorithm
const (
	X25519    = multicodec.X25519Pub
	Ed25519   = multicodec.Ed25519Pub // UCAN required/recommended
	P256      = multicodec.P256Pub    // UCAN required
	P384      = multicodec.P384Pub
	P521      = multicodec.P521Pub
	Secp256k1 = multicodec.Secp256k1Pub // UCAN required
	RSA       = multicodec.RsaPub
)

// Undef can be used to represent a nil or undefined DID, using DID{}
// directly is also acceptable.
var Undef = DID{}

// DID is a Decentralized Identifier of the did:key type, directly holding a cryptographic public key.
// [did:key format]: https://w3c-ccg.github.io/did-method-key/
type DID struct {
	code  multicodec.Code
	bytes string // as string instead of []byte to allow the == operator
}

func Parse(str string) (DID, error) {
	const keyPrefix = "did:key:"

	if !strings.HasPrefix(str, keyPrefix) {
		return Undef, fmt.Errorf("must start with 'did:key'")
	}

	baseCodec, bytes, err := mbase.Decode(str[len(keyPrefix):])
	if err != nil {
		return Undef, err
	}
	if baseCodec != mbase.Base58BTC {
		return Undef, fmt.Errorf("not Base58BTC encoded")
	}
	code, _, err := varint.FromUvarint(bytes)
	if err != nil {
		return Undef, err
	}
	switch multicodec.Code(code) {
	case Ed25519, P256, Secp256k1, RSA:
		return DID{bytes: string(bytes), code: multicodec.Code(code)}, nil
	default:
		return Undef, fmt.Errorf("unsupported did:key multicodec: 0x%x", code)
	}
}

func MustParse(str string) DID {
	did, err := Parse(str)
	if err != nil {
		panic(err)
	}
	return did
}

// Defined tells if the DID is defined, not equal to Undef.
func (d DID) Defined() bool {
	return d.code == 0 || len(d.bytes) > 0
}

// PubKey returns the public key encapsulated in the did:key.
func (d DID) PubKey() (crypto.PubKey, error) {
	unmarshaler, ok := map[multicodec.Code]crypto.PubKeyUnmarshaller{
		X25519:    crypto.UnmarshalEd25519PublicKey,
		Ed25519:   crypto.UnmarshalEd25519PublicKey,
		P256:      crypto.UnmarshalECDSAPublicKey,
		P384:      crypto.UnmarshalECDSAPublicKey,
		P521:      crypto.UnmarshalECDSAPublicKey,
		Secp256k1: crypto.UnmarshalSecp256k1PublicKey,
		RSA:       crypto.UnmarshalRsaPublicKey,
	}[d.code]
	if !ok {
		return nil, fmt.Errorf("unsupported multicodec: %d", d.code)
	}

	codeSize := varint.UvarintSize(uint64(d.code))
	return unmarshaler([]byte(d.bytes)[codeSize:])
}

// String formats the decentralized identity document (DID) as a string.
func (d DID) String() string {
	key, _ := mbase.Encode(mbase.Base58BTC, []byte(d.bytes))
	return "did:key:" + key
}
