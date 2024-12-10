package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/d-voting/contracts/evoting"
	"go.dedis.ch/d-voting/proxy/txnmanager"
	"go.dedis.ch/dela"
	"go.dedis.ch/dela/contracts/access"
	"go.dedis.ch/dela/core/execution/native"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/core/txn"
	"go.dedis.ch/dela/core/txn/signed"
	"go.dedis.ch/dela/crypto"
	"golang.org/x/xerrors"
)

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

// For scenarioTest
func pollTxnInclusion(maxPollCount int, interPollWait time.Duration, proxyAddr, token string, t *testing.T) (bool, error) {
	t.Logf("Starting polling for transaction inclusion")
	for i := 0; i < maxPollCount; i++ {
		if i%10 == 0 {
			t.Logf("Polling for transaction inclusion: %d/%d", i, maxPollCount)
		}
		timeBegin := time.Now()

		req, err := http.NewRequest(http.MethodGet, proxyAddr+"/evoting/transactions/"+token, bytes.NewBuffer([]byte("")))
		if err != nil {
			return false, xerrors.Errorf("failed to create request: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, xerrors.Errorf("failed retrieve the decryption from the server: %v", err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, xerrors.Errorf("failed to read response body: %v", err)
		}
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected status: %s", body)

		// get the body of the response as json
		var result txnmanager.TransactionClientInfo
		err = json.Unmarshal(body, &result)
		if err != nil {
			return false, xerrors.Errorf("failed to unmarshal response body: %v", err)
		}

		// check if the transaction is included in the blockchain

		switch result.Status {
		case 2:
			return false, nil
		case 1:
			t.Logf("Transaction included in the blockchain at iteration: %d/%d", i, maxPollCount)
			return true, nil
		case 0:
			token = result.Token
		}

		if time.Since(timeBegin) < interPollWait {
			time.Sleep(interPollWait - time.Since(timeBegin))
		}

	}

	return false, xerrors.Errorf("transaction not included after timeout")
}

// For integrationTest
func (m txManager) addAndWait(args ...txn.Arg) ([]byte, error) {
	for i := 0; i < m.retry; i++ {
		dela.Logger.Info().Msgf("Adding and waiting for tx to succeed: %d", i)
		sentTxn, err := m.m.Make(args...)
		if err != nil {
			return nil, xerrors.Errorf("failed to Make: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), m.t)
		defer cancel()

		events := m.n.GetOrdering().Watch(ctx)

		err = m.n.GetPool().Add(sentTxn)
		if err != nil {
			fmt.Printf("Failed to add transaction: %v", err)
			continue
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

		time.Sleep(time.Millisecond * (1 << i))
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
		{Key: access.GrantIDArg, Value: []byte(hex.EncodeToString([]byte(evoting.ContractUID)))},
		{Key: access.GrantContractArg, Value: []byte("go.dedis.ch/dela.Evoting")},
		{Key: access.GrantCommandArg, Value: []byte("all")},
		{Key: access.IdentityArg, Value: []byte(base64.StdEncoding.EncodeToString(pubKeyBuf))},
		{Key: access.CmdArg, Value: []byte(access.CmdSet)},
	}
	_, err = m.addAndWait(args...)
	if err != nil {
		return xerrors.Errorf("failed to grantAccess: %v", err)
	}

	return nil
}
