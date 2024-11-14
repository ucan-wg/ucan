package container

import (
	"encoding/base64"
	"fmt"
	"io"
	"iter"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/datamodel"

	"github.com/ucan-wg/go-ucan/token"
	"github.com/ucan-wg/go-ucan/token/delegation"
	"github.com/ucan-wg/go-ucan/token/invocation"
)

var ErrNotFound = fmt.Errorf("not found")

// Reader is a token container reader. It exposes the tokens conveniently decoded.
type Reader map[cid.Cid]token.Token

// GetToken returns an arbitrary decoded token, from its CID.
// If not found, ErrNotFound is returned.
func (ctn Reader) GetToken(cid cid.Cid) (token.Token, error) {
	tkn, ok := ctn[cid]
	if !ok {
		return nil, ErrNotFound
	}
	return tkn, nil
}

// GetDelegation is the same as GetToken but only return a delegation.Token, with the right type.
func (ctn Reader) GetDelegation(cid cid.Cid) (*delegation.Token, error) {
	tkn, err := ctn.GetToken(cid)
	if err != nil {
		return nil, err
	}
	if tkn, ok := tkn.(*delegation.Token); ok {
		return tkn, nil
	}
	return nil, fmt.Errorf("not a delegation token")
}

// GetAllDelegations returns all the delegation.Token in the container.
func (ctn Reader) GetAllDelegations() iter.Seq2[cid.Cid, *delegation.Token] {
	return func(yield func(cid.Cid, *delegation.Token) bool) {
		for c, t := range ctn {
			if t, ok := t.(*delegation.Token); ok {
				if !yield(c, t) {
					return
				}
			}
		}
	}
}

// GetInvocation returns the first found invocation.Token.
// If none are found, ErrNotFound is returned.
func (ctn Reader) GetInvocation() (*invocation.Token, error) {
	for _, t := range ctn {
		if inv, ok := t.(*invocation.Token); ok {
			return inv, nil
		}
	}
	return nil, ErrNotFound
}

func FromCar(r io.Reader) (Reader, error) {
	_, it, err := readCar(r)
	if err != nil {
		return nil, err
	}

	ctn := make(Reader)

	for block, err := range it {
		if err != nil {
			return nil, err
		}

		err = ctn.addToken(block.data)
		if err != nil {
			return nil, err
		}
	}

	return ctn, nil
}

func FromCarBase64(r io.Reader) (Reader, error) {
	return FromCar(base64.NewDecoder(base64.StdEncoding, r))
}

func FromCbor(r io.Reader) (Reader, error) {
	n, err := ipld.DecodeStreaming(r, dagcbor.Decode)
	if err != nil {
		return nil, err
	}
	if n.Kind() != datamodel.Kind_Map {
		return nil, fmt.Errorf("invalid container format: expected map")
	}
	if n.Length() != 1 {
		return nil, fmt.Errorf("invalid container format: expected single version key")
	}

	// get the first (and only) key-value pair
	it := n.MapIterator()
	key, tokensNode, err := it.Next()
	if err != nil {
		return nil, err
	}

	version, err := key.AsString()
	if err != nil {
		return nil, fmt.Errorf("invalid container format: version must be string")
	}
	if version != currentContainerVersion {
		return nil, fmt.Errorf("unsupported container version: %s", version)
	}

	if tokensNode.Kind() != datamodel.Kind_List {
		return nil, fmt.Errorf("invalid container format: tokens must be a list")
	}

	ctn := make(Reader, tokensNode.Length())
	it2 := tokensNode.ListIterator()
	for !it2.Done() {
		_, val, err := it2.Next()
		if err != nil {
			return nil, err
		}
		data, err := val.AsBytes()
		if err != nil {
			return nil, err
		}
		err = ctn.addToken(data)
		if err != nil {
			return nil, err
		}
	}
	return ctn, nil
}

func FromCborBase64(r io.Reader) (Reader, error) {
	return FromCbor(base64.NewDecoder(base64.StdEncoding, r))
}

func (ctn Reader) addToken(data []byte) error {
	tkn, c, err := token.FromSealed(data)
	if err != nil {
		return err
	}
	ctn[c] = tkn
	return nil
}
