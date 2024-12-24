package api

type StatusLike struct {
	Catchpoint                    *string `json:"catchpoint,omitempty"`
	CatchpointAcquiredBlocks      *int    `json:"catchpoint-acquired-blocks,omitempty"`
	CatchpointProcessedAccounts   *int    `json:"catchpoint-processed-accounts,omitempty"`
	CatchpointProcessedKvs        *int    `json:"catchpoint-processed-kvs,omitempty"`
	CatchpointTotalAccounts       *int    `json:"catchpoint-total-accounts,omitempty"`
	CatchpointTotalBlocks         *int    `json:"catchpoint-total-blocks,omitempty"`
	CatchpointTotalKvs            *int    `json:"catchpoint-total-kvs,omitempty"`
	CatchpointVerifiedAccounts    *int    `json:"catchpoint-verified-accounts,omitempty"`
	CatchpointVerifiedKvs         *int    `json:"catchpoint-verified-kvs,omitempty"`
	CatchupTime                   int     `json:"catchup-time"`
	LastCatchpoint                *string `json:"last-catchpoint,omitempty"`
	LastRound                     int     `json:"last-round"`
	LastVersion                   string  `json:"last-version"`
	NextVersion                   string  `json:"next-version"`
	NextVersionRound              int     `json:"next-version-round"`
	NextVersionSupported          bool    `json:"next-version-supported"`
	StoppedAtUnsupportedRound     bool    `json:"stopped-at-unsupported-round"`
	TimeSinceLastRound            int     `json:"time-since-last-round"`
	UpgradeDelay                  *int    `json:"upgrade-delay,omitempty"`
	UpgradeNextProtocolVoteBefore *int    `json:"upgrade-next-protocol-vote-before,omitempty"`
	UpgradeNoVotes                *int    `json:"upgrade-no-votes,omitempty"`
	UpgradeNodeVote               *bool   `json:"upgrade-node-vote,omitempty"`
	UpgradeVoteRounds             *int    `json:"upgrade-vote-rounds,omitempty"`
	UpgradeVotes                  *int    `json:"upgrade-votes,omitempty"`
	UpgradeVotesRequired          *int    `json:"upgrade-votes-required,omitempty"`
	UpgradeYesVotes               *int    `json:"upgrade-yes-votes,omitempty"`
}
