package participation

import (
	"encoding/base64"
	"github.com/algorandfoundation/nodekit/api"
	"testing"
)

func Test_IntegrityHash(t *testing.T) {
	account := "TUIDKH2C7MUHZDD77MAMUREJRKNK25SYXB7OAFA6JFBB24PEL5UX4S4GUU"
	selectionKey, err := base64.StdEncoding.DecodeString("DM9cyZ0oLuVHDtVzhkhLIW06uE0J9faf6aL/FeLFj3o=")
	if err != nil {
		t.Fatal(err)
	}
	stateProofKey, err := base64.StdEncoding.DecodeString("+DAZBTXOletJxFUhEaYQaWaNs3Q4DLEwOlJ68gI8IGq9Ss/1szOimQiAt+f6lqk4FxEe/XvaAXkMbv2/9OiE1g==")
	if err != nil {
		t.Fatal(err)
	}

	voteKey, err := base64.StdEncoding.DecodeString("dHynahCuNWpeR9BcE+B8VE1GM/KdUj759k9ja8zNY30=")
	if err != nil {
		t.Fatal(err)
	}

	var fv = 47733256
	var lv = 47861731
	var kd = 359

	var expectedHash = "4OAJOXKPLUQM2"

	res, err := IntegrityHash(api.ParticipationKey{
		Address:             account,
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "",
		Key: api.AccountParticipation{
			SelectionParticipationKey: selectionKey,
			StateProofKey:             &stateProofKey,
			VoteFirstValid:            fv,
			VoteKeyDilution:           kd,
			VoteLastValid:             lv,
			VoteParticipationKey:      voteKey,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	})
	if res != expectedHash {
		t.Error("expected", expectedHash, "got", res)
	}

	var expectedOfflineHash = "FFDBPNDG63S7C"
	res, err = OfflineHash(account, "testnet")
	if res != expectedOfflineHash {
		t.Error("expected", expectedOfflineHash, "got", res)
	}

}
