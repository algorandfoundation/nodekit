package api

// StatusLike is the anonymous JSON200 struct that oapi-codegen emits for both
// GetStatusResponse and WaitForBlockResponse, so the two can be merged through
// one shape and mocks can construct one without restating it.
//
// It is deliberately an alias rather than a named type: that makes it the same
// type as the generated fields, so values pass either way with no conversion.
// It must stay field-for-field identical to the generated struct -- if
// `make generate` picks up a newer algod spec and the compiler complains here,
// re-copy the JSON200 body out of api/lf.go.
type StatusLike = struct {
	// Catchpoint The current catchpoint that is being caught up to
	Catchpoint *string `json:"catchpoint,omitempty"`

	// CatchpointAcquiredBlocks The number of blocks that have already been obtained by the node as part of the catchup
	CatchpointAcquiredBlocks *uint64 `json:"catchpoint-acquired-blocks,omitempty"`

	// CatchpointProcessedAccounts The number of accounts from the current catchpoint that have been processed so far as part of the catchup
	CatchpointProcessedAccounts *uint64 `json:"catchpoint-processed-accounts,omitempty"`

	// CatchpointProcessedKvs The number of key-values (KVs) from the current catchpoint that have been processed so far as part of the catchup
	CatchpointProcessedKvs *uint64 `json:"catchpoint-processed-kvs,omitempty"`

	// CatchpointTotalAccounts The total number of accounts included in the current catchpoint
	CatchpointTotalAccounts *uint64 `json:"catchpoint-total-accounts,omitempty"`

	// CatchpointTotalBlocks The total number of blocks that are required to complete the current catchpoint catchup
	CatchpointTotalBlocks *uint64 `json:"catchpoint-total-blocks,omitempty"`

	// CatchpointTotalKvs The total number of key-values (KVs) included in the current catchpoint
	CatchpointTotalKvs *uint64 `json:"catchpoint-total-kvs,omitempty"`

	// CatchpointVerifiedAccounts The number of accounts from the current catchpoint that have been verified so far as part of the catchup
	CatchpointVerifiedAccounts *uint64 `json:"catchpoint-verified-accounts,omitempty"`

	// CatchpointVerifiedKvs The number of key-values (KVs) from the current catchpoint that have been verified so far as part of the catchup
	CatchpointVerifiedKvs *uint64 `json:"catchpoint-verified-kvs,omitempty"`

	// CatchupTime CatchupTime in nanoseconds
	CatchupTime int64 `json:"catchup-time"`

	// LastCatchpoint The last catchpoint seen by the node
	LastCatchpoint *string `json:"last-catchpoint,omitempty"`

	// LastRound LastRound indicates the last round seen
	LastRound uint64 `json:"last-round"`

	// LastVersion LastVersion indicates the last consensus version supported
	LastVersion string `json:"last-version"`

	// NextVersion NextVersion of consensus protocol to use
	NextVersion string `json:"next-version"`

	// NextVersionRound NextVersionRound is the round at which the next consensus version will apply
	NextVersionRound uint64 `json:"next-version-round"`

	// NextVersionSupported NextVersionSupported indicates whether the next consensus version is supported by this node
	NextVersionSupported bool `json:"next-version-supported"`

	// StoppedAtUnsupportedRound StoppedAtUnsupportedRound indicates that the node does not support the new rounds and has stopped making progress
	StoppedAtUnsupportedRound bool `json:"stopped-at-unsupported-round"`

	// TimeSinceLastRound TimeSinceLastRound in nanoseconds
	TimeSinceLastRound int64 `json:"time-since-last-round"`

	// UpgradeDelay Upgrade delay
	UpgradeDelay *uint64 `json:"upgrade-delay,omitempty"`

	// UpgradeNextProtocolVoteBefore Next protocol round
	UpgradeNextProtocolVoteBefore *uint64 `json:"upgrade-next-protocol-vote-before,omitempty"`

	// UpgradeNoVotes No votes cast for consensus upgrade
	UpgradeNoVotes *uint64 `json:"upgrade-no-votes,omitempty"`

	// UpgradeNodeVote This node's upgrade vote
	UpgradeNodeVote *bool `json:"upgrade-node-vote,omitempty"`

	// UpgradeVoteRounds Total voting rounds for current upgrade
	UpgradeVoteRounds *uint64 `json:"upgrade-vote-rounds,omitempty"`

	// UpgradeVotes Total votes cast for consensus upgrade
	UpgradeVotes *uint64 `json:"upgrade-votes,omitempty"`

	// UpgradeVotesRequired Yes votes required for consensus upgrade
	UpgradeVotesRequired *uint64 `json:"upgrade-votes-required,omitempty"`

	// UpgradeYesVotes Yes votes cast for consensus upgrade
	UpgradeYesVotes *uint64 `json:"upgrade-yes-votes,omitempty"`
}
