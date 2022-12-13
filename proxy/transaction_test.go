package proxy

import (
	"encoding/hex"
	"net/http"
	"testing"

	//	"github.com/dedis/d-voting/contracts/evoting"
	//"github.com/dedis/d-voting/contracts/evoting/types"
	//"github.com/dedis/d-voting/internal/testing/fake"
	//	ptypes "github.com/dedis/d-voting/proxy/types"
	"go.dedis.ch/dela/core/access"
	"go.dedis.ch/dela/serde"
	"go.dedis.ch/kyber/v3"
	//	"github.com/stretchr/testify/mock"
	//	"com.zerolog.logger"
)



func IsTxnIncludedTest(t *testing.T, w http.ResponseWriter, r *http.Request) {
	// TODO
///	h := initForm()

	//creation of the txn
	//txnID, lastBlock, err := h.submitTxn(r.Context(), evoting.CmdCastVote, evoting.FormArg, data)
	//if err != nil {
//		http.Error(w, "failed to submit txn: "+err.Error(), http.StatusInternalServerError)
//		return
//	}

	//send the transaction
//	h.sendTransactionInfo(w, txnID, lastBlock, ptypes.UnknownTransactionStatus)




}


//------------------------------------MOCK CREATION------------------------------------//

type fakeDKG struct {
	actor fakeDkgActor
	err   error
}

type fakeDkgActor struct {
	publicKey kyber.Point
	err       error
}

var dummyFormIDBuff = []byte("dummyID")
var fakeFormID = hex.EncodeToString(dummyFormIDBuff)

type fakeAccess struct {
	// A package that is used to access the blockchain.
	access.Service

	err error
}

type fakeAuthorityFactory struct {
	serde.Factory
}
/*defining the behavior of the mock logger
// create a mock zeroLog.Logger
var mockLogger = new(mockLogger.MockLogger)
var mockContext = new(mockContext.MockContext)
var mockFactory = new(mockFactory.MockFactory)

func main() {
	// define the behavior of the mock logger
	mockLogger.On("IsTraceEnabled").Return(true)
	mockLogger.On("IsDebugEnabled").Return(true)
	mockLogger.On("IsInfoEnabled").Return(true)
	mockLogger.On("IsWarnEnabled").Return(true)
	mockLogger.On("IsErrorEnabled").Return(true)
}


func initForm( mockLogger logger, mockContext serde.Context, mockFactory serde.Factory) (types.Form) {
	fakeDkg := fakeDKG{
		actor: fakeDkgActor{},
		err:   nil,
	}

	//real

	dummyForm := form{
		sync.Mutex : sync.Mutex{},

		orderingSvc : fake.OrderingService{},
		logger      : mockLogger,
		context     : mockContext,
		formFac     : mockFactory,
		mngr        txn.Manager
		pool        pool.Pool
		pk          kyber.Point
		blocks      blockstore.BlockStore
		signer      crypto.Signer
		

		FormID:           fakeFormID,
		Status:           0,
		Pubkey:           nil,
		Suffragia:        types.Suffragia{},
		ShuffleInstances: make([]types.ShuffleInstance, 0),
		DecryptedBallots: nil,
		ShuffleThreshold: 0,
		Roster:           fake.Authority{},
	}

	//fake

	var evotingAccessKey = [32]byte{3}
	rosterKey := [32]byte{}

	service := fakeAccess{err: fake.GetError()}
	rosterFac := fakeAuthorityFactory{}

	return dummyForm
}
*/