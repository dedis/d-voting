package fake

import (
	"io"

	"go.dedis.ch/dela/core/ordering/cosipbft/authority"
	"go.dedis.ch/dela/mino"
	"go.dedis.ch/dela/serde"
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

// ChangeSet is a fake implementation of ordering.cosipbft.authority.ChangeSetFactory.
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

// Taken from Dela
/* var rosterFormats = registry.NewSimpleRegistry()

// RegisterRosterFormat registers the engine for the provided format.
func RegisterRosterFormat(c serde.Format, f serde.FormatEngine) {
	rosterFormats.Register(c, f)
}

// iterator is a generic implementation of an iterator over a list of conodes.
type iterator struct {
	index  int
	roster *Roster
}

func (i *iterator) Seek(index int) {
	i.index = index
}

func (i *iterator) HasNext() bool {
	return i.index < i.roster.Len()
}

func (i *iterator) GetNext() int {
	res := i.index
	i.index++
	return res
}

// addressIterator is an iterator for a list of addresses.
//
// - implements mino.AddressIterator
type addressIterator struct {
	*iterator
}

// GetNext implements mino.AddressIterator. It returns the next address.
func (i *addressIterator) GetNext() mino.Address {
	if i.iterator.HasNext() {
		return i.roster.addrs[i.iterator.GetNext()]
	}
	return nil
}

// publicKeyIterator is an iterator for a list of public keys.
//
// - implements crypto.PublicKeyIterator
type publicKeyIterator struct {
	*iterator
}

// GetNext implements crypto.PublicKeyIterator. It returns the next public key.
func (i *publicKeyIterator) GetNext() crypto.PublicKey {
	if i.iterator.HasNext() {
		return i.roster.pubkeys[i.iterator.GetNext()]
	}
	return nil
}

// Roster contains a list of participants with their addresses and public keys.
//
// - implements authority.Authority
type Roster struct {
	addrs   []mino.Address
	pubkeys []crypto.PublicKey
}

// New creates a new roster from the list of addresses and public keys.
func New(addrs []mino.Address, pubkeys []crypto.PublicKey) Roster {
	return Roster{
		addrs:   addrs,
		pubkeys: pubkeys,
	}
}

// FromAuthority returns a viewchange roster from a collective authority.
func FromAuthority(authority crypto.CollectiveAuthority) Roster {
	addrs := make([]mino.Address, authority.Len())
	pubkeys := make([]crypto.PublicKey, authority.Len())

	addrIter := authority.AddressIterator()
	pubkeyIter := authority.PublicKeyIterator()
	for i := 0; addrIter.HasNext() && pubkeyIter.HasNext(); i++ {
		addrs[i] = addrIter.GetNext()
		pubkeys[i] = pubkeyIter.GetNext()
	}

	return New(addrs, pubkeys)
}

// Fingerprint implements serde.Fingerprinter. It marshals the roster and writes
// the result in the given writer.
func (r Roster) Fingerprint(w io.Writer) error {
	for i, addr := range r.addrs {
		data, err := addr.MarshalText()
		if err != nil {
			return xerrors.Errorf("couldn't marshal address: %v", err)
		}

		_, err = w.Write(data)
		if err != nil {
			return xerrors.Errorf("couldn't write address: %v", err)
		}

		data, err = r.pubkeys[i].MarshalBinary()
		if err != nil {
			return xerrors.Errorf("couldn't marshal public key: %v", err)
		}

		_, err = w.Write(data)
		if err != nil {
			return xerrors.Errorf("couldn't write public key: %v", err)
		}
	}

	return nil
}

// Take implements mino.Players. It returns a subset of the roster according to
// the filter.
func (r Roster) Take(updaters ...mino.FilterUpdater) mino.Players {
	filter := mino.ApplyFilters(updaters)
	newRoster := Roster{
		addrs:   make([]mino.Address, len(filter.Indices)),
		pubkeys: make([]crypto.PublicKey, len(filter.Indices)),
	}

	for i, k := range filter.Indices {
		newRoster.addrs[i] = r.addrs[k]
		newRoster.pubkeys[i] = r.pubkeys[k]
	}

	return newRoster
}


// Apply implements authority.Authority. It returns a new authority after
// applying the change set. The removals must be sorted by descending order and
// unique or the behaviour will be undefined.
func (r Roster) Apply(in authority.ChangeSet) authority.Authority {
	changeset, ok := in.(*RosterChangeSet)
	if !ok {
		dela.Logger.Warn().Msgf("Change set '%T' is not supported. Ignoring.", in)
		return r
	}

	addrs := make([]mino.Address, r.Len())
	pubkeys := make([]crypto.PublicKey, r.Len())

	for i, addr := range r.addrs {
		addrs[i] = addr
		pubkeys[i] = r.pubkeys[i]
	}

	for _, i := range changeset.remove {
		if int(i) < len(addrs) {
			addrs = append(addrs[:i], addrs[i+1:]...)
			pubkeys = append(pubkeys[:i], pubkeys[i+1:]...)
		}
	}

	roster := Roster{
		addrs:   append(addrs, changeset.addrs...),
		pubkeys: append(pubkeys, changeset.pubkeys...),
	}

	return roster
}

// Diff implements authority.Authority. It returns the change set that must be
// applied to the current authority to get the given one.
func (r Roster) Diff(o authority.Authority) ChangeSet {
	changeset := NewChangeSet()

	other, ok := o.(Roster)
	if !ok {
		return changeset
	}

	i := 0
	k := 0
	for i < len(r.addrs) || k < len(other.addrs) {
		if i < len(r.addrs) && k < len(other.addrs) {
			if r.addrs[i].Equal(other.addrs[k]) {
				i++
				k++
			} else {
				changeset.remove = append(changeset.remove, uint(i))
				i++
			}
		} else if i < len(r.addrs) {
			changeset.remove = append(changeset.remove, uint(i))
			i++
		} else {
			changeset.addrs = append(changeset.addrs, other.addrs[k])
			changeset.pubkeys = append(changeset.pubkeys, other.pubkeys[k])
			k++
		}
	}

	return changeset
}

// Len implements mino.Players. It returns the length of the authority.
func (r Roster) Len() int {
	return len(r.addrs)
}

// GetPublicKey implements crypto.CollectiveAuthority. It returns the public key
// of the address if it exists, nil otherwise. The second return is the index of
// the public key in the authority.
func (r Roster) GetPublicKey(target mino.Address) (crypto.PublicKey, int) {
	for i, addr := range r.addrs {
		if addr.Equal(target) {
			return r.pubkeys[i], i
		}
	}

	return nil, -1
}

// AddressIterator implements mino.Players. It returns an iterator of the
// addresses of the roster in a deterministic order.
func (r Roster) AddressIterator() mino.AddressIterator {
	return &addressIterator{iterator: &iterator{roster: &r}}
}

// PublicKeyIterator implements crypto.CollectiveAuthority. It returns an
// iterator of the public keys of the roster in a deterministic order.
func (r Roster) PublicKeyIterator() crypto.PublicKeyIterator {
	return &publicKeyIterator{iterator: &iterator{roster: &r}}
}

// Serialize implements serde.Message. It returns the serialized data for this
// roster.
func (r Roster) Serialize(ctx serde.Context) ([]byte, error) {
	format := rosterFormats.Get(ctx.GetFormat())

	data, err := format.Encode(ctx, r)
	if err != nil {
		return nil, xerrors.Errorf("couldn't encode roster: %v", err)
	}

	return data, nil
}

// rosterFac is a factory to deserialize authority.
//
// - implements authority.Factory
type rosterFac struct {
	addrFactory   mino.AddressFactory
	pubkeyFactory crypto.PublicKeyFactory
}

// NewFactory creates a new instance of the authority factory.
func NewFactory(af mino.AddressFactory, pf crypto.PublicKeyFactory) authority.Factory {
	return rosterFac{
		addrFactory:   af,
		pubkeyFactory: pf,
	}
}

// Deserialize implements serde.Factory.  It returns the roster from the data if
// appropriate, otherwise an error.
func (f rosterFac) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	return f.AuthorityOf(ctx, data)
}

// AuthorityOf implements authority.AuthorityFactory. It returns the roster from
// the data if appropriate, otherwise an error.
func (f rosterFac) AuthorityOf(ctx serde.Context, data []byte) (authority.Authority, error) {
	format := rosterFormats.Get(ctx.GetFormat())

	ctx = serde.WithFactory(ctx, PubKeyFac{}, f.pubkeyFactory)
	ctx = serde.WithFactory(ctx, AddrKeyFac{}, f.addrFactory)

	msg, err := format.Decode(ctx, data)
	if err != nil {
		return nil, xerrors.Errorf("couldn't decode roster: %v", err)
	}

	roster, ok := msg.(Roster)
	if !ok {
		return nil, xerrors.Errorf("invalid message of type '%T'", msg)
	}

	return roster, nil
} */
