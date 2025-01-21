package participation

import (
	"crypto/sha512"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/nodekit/api"
	"strings"
)

// concat merges multiple byte slices into a single byte slice by appending all input slices into a single output slice.
func concat(slices [][]byte) []byte {
	var all []byte
	for _, s := range slices {
		all = append(all, s...)
	}
	return all
}

// hash calculates a shortened base32-encoded SHA-512/256 hash of the input byte slice and returns it as a string.
func hash(rawBytes []byte) string {
	hashBytes := sha512.Sum512_256(rawBytes)
	return strings.Replace(base32.StdEncoding.EncodeToString(hashBytes[:8]), "===", "", 1)
}

// IntegrityHash computes a unique, deterministic hash based on a ParticipationKey, ensuring data integrity for validation.
// Returns a base32-encoded string and an error if the input data is invalid or hashing fails.
func IntegrityHash(partkey api.ParticipationKey) (string, error) {
	address, err := types.DecodeAddress(partkey.Address)
	if err != nil {
		return "", err
	}

	encodedFV := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedFV, uint64(partkey.Key.VoteFirstValid))

	encodedLV := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedLV, uint64(partkey.Key.VoteLastValid))

	encodedVoteKeyDilution := make([]byte, 8)
	binary.BigEndian.PutUint64(encodedVoteKeyDilution, uint64(partkey.Key.VoteKeyDilution))

	// raw bytes
	rawData := concat([][]byte{
		address[:],
		partkey.Key.SelectionParticipationKey[:],
		partkey.Key.VoteParticipationKey[:],
		*partkey.Key.StateProofKey,
		encodedFV[:],
		encodedLV[:],
		encodedVoteKeyDilution[:],
	})

	if len(rawData) != 184 {
		return "", errors.New("invalid raw data length")
	}

	// Enchode
	hashBytes := sha512.Sum512_256(rawData)
	return strings.Replace(base32.StdEncoding.EncodeToString(hashBytes[:8]), "===", "", 1), nil
}

func OfflineHash(address string, network string) (string, error) {
	addr, err := types.DecodeAddress(address)
	if err != nil {
		return "", err
	}

	rawBytes := concat([][]byte{
		addr[:],
		[]byte(network),
	})

	return hash(rawBytes), nil
}
