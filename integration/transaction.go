package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"time"

	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"golang.org/x/xerrors"
)

const addAndWaitErr = "failed to addAndWait: %v"

func newTxManager(signer crypto.Signer, firstNode dVotingCosiDela,
	timeout time.Duration, retry int) txManager {

	client := client{
		srvc: firstNode.GetOrdering(),
		mgr:  firstNode.GetValidationSrv(),
	}

	return txManager{
		m:     signed.NewManager(signer, client),
		n:     firstNode,
		t:     timeout,
		retry: retry,
	}
}

type txManager struct {
	m     txn.Manager
	n     dVotingCosiDela
	t     time.Duration
	retry int
}

func (m txManager) addAndWait(args ...txn.Arg) ([]byte, error) {
	for i := 0; i < m.retry; i++ {
		sentTxn, err := m.m.Make(args...)
		if err != nil {
			return nil, xerrors.Errorf("failed to Make: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), m.t)
		defer cancel()

		events := m.n.GetOrdering().Watch(ctx)

		err = m.n.GetPool().Add(sentTxn)
		if err != nil {
			return nil, xerrors.Errorf("failed to Add: %v", err)
		}

		sentTxnID := sentTxn.GetID()

		accepted := isAccepted(events, sentTxnID)
		if accepted {
			return sentTxnID, nil
		}

		err = m.m.Sync()
		if err != nil {
			return nil, xerrors.Errorf("failed to sync: %v", err)
		}

		cancel()
	}

	return nil, xerrors.Errorf("transaction not included after timeout: %v", args)
}

// isAccepted returns true if the transaction was included then accepted
func isAccepted(events <-chan ordering.Event, txID []byte) bool {
	for event := range events {
		for _, result := range event.Transactions {
			fetchedTxnID := result.GetTransaction().GetID()

			if bytes.Equal(txID, fetchedTxnID) {
				accepted, _ := event.Transactions[0].GetStatus()

				return accepted
			}
		}
	}

	return false
}

func grantAccess(m txManager, signer crypto.Signer) error {
	pubKeyBuf, err := signer.GetPublicKey().MarshalBinary()
	if err != nil {
		return xerrors.Errorf("failed to GetPublicKey: %v", err)
	}

	args := []txn.Arg{
		{Key: native.ContractArg, Value: []byte("go.dedis.ch/dela.Access")},
		{Key: "access:grant_id", Value: []byte(hex.EncodeToString(evotingAccessKey[:]))},
		{Key: "access:grant_contract", Value: []byte("go.dedis.ch/dela.Evoting")},
		{Key: "access:grant_command", Value: []byte("all")},
		{Key: "access:identity", Value: []byte(base64.StdEncoding.EncodeToString(pubKeyBuf))},
		{Key: "access:command", Value: []byte("GRANT")},
	}
	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf("failed to grantAccess: %v", err)
	}

	return nil
}
