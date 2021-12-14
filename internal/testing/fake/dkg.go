package fake

import (
	"github.com/dedis/d-voting/services/dkg"
	"go.dedis.ch/kyber/v3"
)

type Pedersen struct {
	Err error
}

func (f Pedersen) Listen(electionID []byte) (dkg.Actor, error) {
	return nil, f.Err
}

func (f Pedersen) GetActor(electionID []byte) (dkg.Actor, bool) {
	return nil, false
}

type DKGActor struct {
	Err    error
	PubKey kyber.Point
}

func (f DKGActor) Setup() (pubKey kyber.Point, err error) {
	return f.PubKey, f.Err
}

func (f DKGActor) GetPublicKey() (kyber.Point, error) {
	return f.PubKey, f.Err
}

func (f DKGActor) Encrypt(message []byte) (K, C kyber.Point, remainder []byte, err error) {
	return nil, nil, nil, f.Err
}

func (f DKGActor) Decrypt(K, C kyber.Point) ([]byte, error) {
	return nil, f.Err
}

func (f DKGActor) Reshare() error {
	return f.Err
}

func (f DKGActor) MarshalJSON() ([]byte, error) {
	return nil, f.Err
}

type PedersenStore struct {
	Err error
	DB  InMemoryDB
}

func (f PedersenStore) DKGStore() {}
