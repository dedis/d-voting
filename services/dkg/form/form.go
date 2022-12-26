package form

import (
	"encoding/hex"

	etypes "github.com/dedis/d-voting/contracts/evoting/types"
	"go.dedis.ch/dela/core/ordering"
	"go.dedis.ch/dela/serde"
	"golang.org/x/xerrors"
)

//add the unique implementation of the getForm method
func getForm(formFac serde.Factory, ctx serde.Context, formIDHex string, srv ordering.Service) (etypes.Form, error) {
	var form etypes.Form //unused

	formID, err := hex.DecodeString(formIDHex)
	if err != nil {
		return form, xerrors.Errorf("failed to decode formIDHex: %v", err)
	}

	proof, err := srv.GetProof(formID)
	if err != nil {
		return form, xerrors.Errorf("failed to get proof: %v", err)
	}

	if string(proof.GetValue()) == "" {
		return form, xerrors.Errorf("form does not exist")
	}

	message, err := formFac.Deserialize(ctx, proof.GetValue())
	if err != nil {
		return form, xerrors.Errorf("failed to deserialize Form: %v", err)
	}

	form, ok := message.(etypes.Form)
	if !ok {
		return form, xerrors.Errorf("wrong message type: %T", message)
	}

	if formIDHex != form.FormID {
		return form, xerrors.Errorf("formID do not match: %q != %q", formIDHex, form.FormID)
	}

	return form, nil
}
