package fake

import (
	"hash"

	"go.dedis.ch/dela/crypto"
	"go.dedis.ch/dela/serde"
)

// PublicKeyFactory is a fake implementation of a public key factory.
//
// - implements crypto.PublicKeyFactory.
type PublicKeyFactory struct {
	pubkey PublicKey
	err    error
}

// NewPublicKeyFactory returns a fake public key factory that returns the given
// public key.
func NewPublicKeyFactory(pubkey PublicKey) PublicKeyFactory {
	return PublicKeyFactory{
		pubkey: pubkey,
	}
}

// NewBadPublicKeyFactory returns a fake public key factory that returns an
// error when appropriate.
func NewBadPublicKeyFactory() PublicKeyFactory {
	return PublicKeyFactory{err: fakeErr}
}

// Deserialize implements serde.Factory.
func (f PublicKeyFactory) Deserialize(serde.Context, []byte) (serde.Message, error) {
	return f.pubkey, f.err
}

// PublicKeyOf implements crypto.PublicKeyFactory.
func (f PublicKeyFactory) PublicKeyOf(serde.Context, []byte) (crypto.PublicKey, error) {
	return f.pubkey, f.err
}

// FromBytes implements crypto.PublicKeyFactory.
func (f PublicKeyFactory) FromBytes([]byte) (crypto.PublicKey, error) {
	return f.pubkey, f.err
}

// SignatureByte is the byte returned when marshaling a fake signature.
const SignatureByte = 0xfe

// Signature is a fake implementation of the signature.
//
// - implements crypto.Signature
type Signature struct {
	crypto.Signature
	err error
}

// NewBadSignature returns a signature that will return error when appropriate.
func NewBadSignature() Signature {
	return Signature{err: fakeErr}
}

// Equal implements crypto.Signature.
func (s Signature) Equal(o crypto.Signature) bool {
	_, ok := o.(Signature)
	return ok
}

// Serialize implements serde.Message.
func (s Signature) Serialize(serde.Context) ([]byte, error) {
	return []byte("{}"), s.err
}

// MarshalBinary implements crypto.Signature.
func (s Signature) MarshalBinary() ([]byte, error) {
	return []byte{SignatureByte}, s.err
}

// String implements fmt.Stringer.
func (s Signature) String() string {
	return "fakeSignature"
}

// SignatureFactory is a fake implementation of the signature factory.
//
// - implements crypto.SignatureFactory
type SignatureFactory struct {
	Counter   *Counter
	signature Signature
	err       error
}

// NewSignatureFactory returns a fake signature factory.
func NewSignatureFactory(s Signature) SignatureFactory {
	return SignatureFactory{signature: s}
}

// NewBadSignatureFactory returns a signature factory that will return an error
// when appropriate.
func NewBadSignatureFactory() SignatureFactory {
	return SignatureFactory{err: fakeErr}
}

// NewBadSignatureFactoryWithDelay returns a signature factory that will return
// an error after some calls.
func NewBadSignatureFactoryWithDelay(value int) SignatureFactory {
	return SignatureFactory{
		err:     fakeErr,
		Counter: &Counter{Value: value},
	}
}

// Deserialize implements serde.Factory.
func (f SignatureFactory) Deserialize(ctx serde.Context, data []byte) (serde.Message, error) {
	return f.SignatureOf(ctx, data)
}

// SignatureOf implements crypto.SignatureFactory.
func (f SignatureFactory) SignatureOf(serde.Context, []byte) (crypto.Signature, error) {
	if !f.Counter.Done() {
		f.Counter.Decrease()
		return f.signature, nil
	}
	return f.signature, f.err
}

// PublicKey is a fake implementation of crypto.PublicKey.
//
// - implements crypto.PublicKey
type PublicKey struct {
	crypto.PublicKey
	err       error
	verifyErr error
}

// NewBadPublicKey returns a new fake public key that returns error when
// appropriate.
func NewBadPublicKey() PublicKey {
	return PublicKey{
		err:       fakeErr,
		verifyErr: fakeErr,
	}
}

// NewInvalidPublicKey returns a fake public key that never verifies.
func NewInvalidPublicKey() PublicKey {
	return PublicKey{verifyErr: fakeErr}
}

// Verify implements crypto.PublicKey.
func (pk PublicKey) Verify([]byte, crypto.Signature) error {
	return pk.verifyErr
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (pk PublicKey) MarshalBinary() ([]byte, error) {
	return []byte("PK"), pk.err
}

// MarshalText implements encoding.TextMarshaler.
func (pk PublicKey) MarshalText() ([]byte, error) {
	return pk.MarshalBinary()
}

// Equal implements crypto.PublicKey.
func (pk PublicKey) Equal(other interface{}) bool {
	_, ok := other.(PublicKey)
	return ok
}

// Serialize implements serde.Message.
func (pk PublicKey) Serialize(serde.Context) ([]byte, error) {
	return []byte(`{}`), pk.err
}

// String implements fmt.Stringer.
func (pk PublicKey) String() string {
	return "fake.PublicKey"
}

// Signer is a fake implementation of the crypto.AggregateSigner interface.
//
// - implements crypto.Signer
type Signer struct {
	crypto.AggregateSigner
	publicKey        PublicKey
	signatureFactory SignatureFactory
	verifierFactory  VerifierFactory
	err              error
}

// NewSigner returns a new instance of the fake signer.
func NewSigner() crypto.Signer {
	return Signer{}
}

// NewAggregateSigner returns a new signer that implements aggregation.
func NewAggregateSigner() Signer {
	return Signer{}
}

// NewSignerWithSignatureFactory returns a fake signer with the provided
// factory.
func NewSignerWithSignatureFactory(f SignatureFactory) Signer {
	return Signer{signatureFactory: f}
}

// NewSignerWithVerifierFactory returns a new fake signer with the specific
// verifier factory.
func NewSignerWithVerifierFactory(f VerifierFactory) Signer {
	return Signer{verifierFactory: f}
}

// NewSignerWithPublicKey returns a new fake signer with the specific public
// key.
func NewSignerWithPublicKey(k PublicKey) Signer {
	return Signer{publicKey: k}
}

// NewBadSigner returns a fake signer that will return an error when
// appropriate.
func NewBadSigner() Signer {
	return Signer{err: fakeErr}
}

// GetPublicKeyFactory implements crypto.Signer.
func (s Signer) GetPublicKeyFactory() crypto.PublicKeyFactory {
	return PublicKeyFactory{}
}

// GetSignatureFactory implements crypto.Signer.
func (s Signer) GetSignatureFactory() crypto.SignatureFactory {
	return s.signatureFactory
}

// GetVerifierFactory implements crypto.Signer.
func (s Signer) GetVerifierFactory() crypto.VerifierFactory {
	return s.verifierFactory
}

// GetPublicKey implements crypto.Signer.
func (s Signer) GetPublicKey() crypto.PublicKey {
	return s.publicKey
}

// Sign implements crypto.Signer.
func (s Signer) Sign([]byte) (crypto.Signature, error) {
	return Signature{}, s.err
}

// Aggregate implements crypto.AggregateSigner.
func (s Signer) Aggregate(...crypto.Signature) (crypto.Signature, error) {
	return Signature{}, s.err
}

// Verifier is a fake implementation of crypto.Verifier.
//
// - implements crypto.Verifier
type Verifier struct {
	crypto.Verifier
	err   error
	count *Counter
}

// NewBadVerifier returns a verifier that will return an error when appropriate.
func NewBadVerifier() Verifier {
	return Verifier{err: fakeErr}
}

// NewBadVerifierWithDelay returns a verifier that will return an error after a
// given delay.
func NewBadVerifierWithDelay(value int) Verifier {
	return Verifier{
		err:   fakeErr,
		count: NewCounter(value),
	}
}

// Verify implements crypto.Verifier.
func (v Verifier) Verify(msg []byte, s crypto.Signature) error {
	if !v.count.Done() {
		v.count.Decrease()
		return nil
	}

	return v.err
}

// VerifierFactory is a fake implementation of crypto.VerifierFactory.
//
// - implements crypto.VerifierFactory
type VerifierFactory struct {
	crypto.VerifierFactory
	verifier Verifier
	err      error
	call     *Call
}

// NewVerifierFactory returns a new fake verifier factory.
func NewVerifierFactory(v Verifier) VerifierFactory {
	return VerifierFactory{verifier: v}
}

// NewVerifierFactoryWithCalls returns a new verifier factory that will register
// the calls.
func NewVerifierFactoryWithCalls(c *Call) VerifierFactory {
	return VerifierFactory{call: c}
}

// NewBadVerifierFactory returns a fake verifier factory that returns an error
// when appropriate.
func NewBadVerifierFactory() VerifierFactory {
	return VerifierFactory{err: fakeErr}
}

// FromAuthority implements crypto.VerifierFactory.
func (f VerifierFactory) FromAuthority(ca crypto.CollectiveAuthority) (crypto.Verifier, error) {
	f.call.Add(ca)

	return f.verifier, f.err
}

// FromArray implements crypto.VerifierFactory.
func (f VerifierFactory) FromArray(pubkeys []crypto.PublicKey) (crypto.Verifier, error) {
	f.call.Add(pubkeys)

	return f.verifier, f.err
}

// Hash is a fake implementation of hash.Hash.
//
// - implements hash.Hash
type Hash struct {
	hash.Hash
	delay int
	err   error
	Call  *Call
}

// NewBadHash returns a fake hash that returns an error when appropriate.
func NewBadHash() *Hash {
	return &Hash{err: fakeErr}
}

// NewBadHashWithDelay returns a fake hash that returns an error after a certain
// amount of calls.
func NewBadHashWithDelay(delay int) *Hash {
	return &Hash{err: fakeErr, delay: delay}
}

// Write implements hash.Hash.
func (h *Hash) Write(in []byte) (int, error) {
	if h.Call != nil {
		h.Call.Add(in)
	}

	if h.delay > 0 {
		h.delay--
		return 0, nil
	}
	return 0, h.err
}

// Size implements hash.Hash.
func (h *Hash) Size() int {
	return 32
}

// Sum implements hash.Hash.
func (h *Hash) Sum([]byte) []byte {
	return make([]byte, 32)
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (h *Hash) MarshalBinary() ([]byte, error) {
	if h.delay > 0 {
		h.delay--
		return []byte{}, nil
	}
	return []byte{}, h.err
}

// UnmarshalBinary implements encodi8ng.BinaryUnmarshaler.
func (h *Hash) UnmarshalBinary([]byte) error {
	if h.delay > 0 {
		h.delay--
		return nil
	}
	return h.err
}

// HashFactory is a fake implementation of a hash factory.
//
// - implements crypto.HashFactory
type HashFactory struct {
	hash *Hash
}

// NewHashFactory returns a fake hash factory.
func NewHashFactory(h *Hash) HashFactory {
	return HashFactory{hash: h}
}

// New implements crypto.HashFactory.
func (f HashFactory) New() hash.Hash {
	return f.hash
}
