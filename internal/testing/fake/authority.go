package fake

import (
	"io"

	"github.com/c4dt/dela/core/ordering/cosipbft/authority"
	"github.com/c4dt/dela/mino"
	"github.com/c4dt/dela/serde"
)

// ChangeSet is a fake implementation of ordering.cosipbft.authority.ChangeSet.
//
// - implements ordering.cosipbft.ChangeSet
type ChangeSet struct {
}

// Serialize implements serde.Message
func (f ChangeSet) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

// NumChanges returns the number of changes that will be applied with this
// change set.
func (f ChangeSet) NumChanges() int {
	return 0
}

// GetNewAddresses returns the list of addresses for the new members.
func (f ChangeSet) GetNewAddresses() []mino.Address {
	return make([]mino.Address, 10)
}

// ChangeSet is a fake implementation of
// ordering.cosipbft.authority.ChangeSetFactory.
//
// - implements ordering.cosipbft.ChangeSetFactory
type ChangeSetFactory struct {
}

// Deserialize implements serde.Factory
func (f ChangeSetFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	return nil, nil
}

func (f ChangeSetFactory) ChangeSetOf(serde.Context, []byte) (authority.ChangeSet, error) {
	return ChangeSet{}, nil
}

// Authority is a fake implementation of ordering.cosipbft.authority.Authority.
//
// - implements ordering.cosipbft.Authority
type Authority struct {
	CollectiveAuthority
}

// Serialize implements serde.Message
func (f Authority) Serialize(ctx serde.Context) ([]byte, error) {
	return nil, nil
}

// Fingerprint implements serde.Fingerprinter
func (f Authority) Fingerprint(writer io.Writer) error {
	return nil
}

// Apply implements ordering.cosipbft.Authority
// Apply must apply the change set to the collective authority. It should
// first remove, then add the new players.
func (f Authority) Apply(authority.ChangeSet) authority.Authority {
	return Authority{}
}

// Diff implements ordering.cosipbft.Authority
// Diff should return the change set to apply to get the given authority.
func (f Authority) Diff(authority.Authority) authority.ChangeSet {
	return ChangeSet{}
}

func (f Authority) Len() int {
	return 1
}

// Factory is a fake implementation of ordering.cosipbft.authority.Factory.
//
// - implements ordering.cosipbft.Factory
type Factory struct {
}

// Deserialize implements serde.Factory
func (f Factory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	return nil, nil
}

// AuthorityOf implements ordering.cosipbft.Factory
func (f Factory) AuthorityOf(serde.Context, []byte) (authority.Authority, error) {
	return Authority{}, nil
}

// This fake RosterFac always returns roster upon Deserialize
type RosterFac struct {
	authority.Factory

	roster authority.Roster
}

func NewRosterFac(roster authority.Roster) RosterFac {
	return RosterFac{roster: roster}
}

func (f RosterFac) AuthorityOf(serde.Context, []byte) (authority.Authority, error) {
	return f.roster, nil
}
