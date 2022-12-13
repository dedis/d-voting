package proxy

import (
	"crypto/cipher"

	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.dedis.ch/kyber/v3"
)



func TestEditShuffle(t *testing.T) {
	// Create a new shuffle instance with a mock actor and public key
	shuffle := NewShuffle(mockActor{}, mockPK{})
  
	// Create a new HTTP request with the "shuffle" action and a valid formID
	req, err := http.NewRequest("POST", "/shuffle", strings.NewReader(`{"action": "shuffle", "formID": "123456"}`))
	if err != nil {
	  t.Fatalf("Error creating request: %v", err)
	}
  
	// Create a new HTTP response recorder to record the response
	rr := httptest.NewRecorder()
  
	// Call the EditShuffle function with the request and response recorder
	shuffle.EditShuffle(rr, req)
  
	// Check the status code of the response
	if status := rr.Code; status != http.StatusOK {
	  t.Errorf("EditShuffle returned wrong status code: got %v want %v",
		status, http.StatusOK)
	}
  
	// Check the response body
	expected := `{"success": true}`
	if rr.Body.String() != expected {
	  t.Errorf("EditShuffle returned unexpected body: got %v want %v",
		rr.Body.String(), expected)
	}
  }

//------------------------------------------------------------MOCKS------------

///////////////////////////////////////////mockActor///////////////////////////

// mockActor is a mock implementation of the shuffleSrv.Actor interface
type mockActor struct{}

// Shuffle is a mock implementation of the shuffleSrv.Actor.Shuffle method
func (a mockActor) Shuffle(formID []byte) error {
  // Return nil to indicate that the shuffle was successful
  return nil
}
///////////////////////////////////////////mockPK//////////////////////////////

type mockPK struct {
	// This field can be used to store any data that the mock public key needs to
	// maintain state, such as the value of the public key.
}

func (m mockPK) Verify(message []byte, signature []byte) bool {
	// This method simulates the behavior of the mock public key when it is called
	// to verify a message and signature. It can return a hard-coded boolean value
	// to indicate whether the verification was successful or not.
	return false
}

func (m mockPK) Add(a kyber.Point, b kyber.Point) kyber.Point{
	// This method simulates the behavior of the mock public key when it is called
	// to add another point to the public key. It can return a hard-coded value to
	// indicate the result of the addition operation.
	return nil

}

func (m mockPK) Base() kyber.Point {
	// This method simulates the behavior of the mock public key when it is called
	// to get the base point of the public key. It can return a hard-coded value to
	// indicate the base point.
	return nil
}
func (m mockPK) Clone() kyber.Point {
	// This method simulates the behavior of the mock public key when it is called
	// to clone the public key. It can return a hard-coded value to indicate the
	// cloned public key.
	return nil
}
func (m mockPK) Data() ([]byte, error) {
	return nil, nil
}
func (m mockPK) Embed(data []byte, r cipher.Stream) kyber.Point {
	return nil

}

func (m mockPK) EmbedLen() int {
	return 0
}
func (m mockPK) MarshalBinary() ([]byte, error) {
	return nil, nil
}
func (m mockPK) MarshalSize() int {
	return 0
}
func (m mockPK) MarshalTo(w io.Writer) (int, error) {
	return 0, nil
}
	

func (m mockPK) Null() kyber.Point {
	return nil
}
func (m mockPK) Equal(b kyber.Point) bool {
	return false //todo
}
func (m mockPK) Pick(rand cipher.Stream) kyber.Point {
	return nil

}
func (m mockPK) Mul(s kyber.Scalar, b kyber.Point) kyber.Point {
	return nil
}
func (m mockPK) Neg(b kyber.Point) kyber.Point {
	return nil
}
func (m mockPK) Sub(a kyber.Point, b kyber.Point) kyber.Point {
	return nil
}
func (m mockPK) Set(a kyber.Point) kyber.Point {
	return nil
}
func (m mockPK) SetInt64(v int64) kyber.Point {
	return nil
}
func (m mockPK) String() string {
	return ""
}
func (m mockPK) UnmarshalBinary(buff []byte) error {
	return nil
}
func (m mockPK) PickLen() int {
	return 0
}
func (m mockPK) PickRand(rand cipher.Stream) kyber.Point {
	return nil
}
func (m mockPK) SetBytes(buff []byte) kyber.Point {
	return nil
}

func (m mockPK) UnmarshalFrom(r io.Reader) (int, error) {
	return 0, nil
}





  
