package fake

import (
	"github.com/dedis/d-voting/services/dkg"
	"go.dedis.ch/kyber/v3"
)

// - implements dkg.DKG
type BadPedersen struct {
	Err error
}

func (f BadPedersen) Listen(electionID []byte) (dkg.Actor, error) {
	return nil, f.Err
}

func (f BadPedersen) GetActor(electionID []byte) (dkg.Actor, bool) {
	return nil, false
}

// - implements dkg.DKG
type Pedersen struct {
	Actors map[string]dkg.Actor
}

func (f Pedersen) Listen(electionID []byte) (dkg.Actor, error) {
	actor := DKGActor{PubKey: suite.Point().Pick(suite.RandomStream())}
	f.Actors[string(electionID)] = actor
	return actor, nil
}

func (f Pedersen) GetActor(electionID []byte) (dkg.Actor, bool) {
	a, exists := f.Actors[string(electionID)]
	return a, exists
}

// - implements dkg.Actor
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
	data, err := f.PubKey.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return data, f.Err
}
