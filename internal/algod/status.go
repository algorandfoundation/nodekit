package algod

import (
	"context"
	"errors"

	"github.com/algorandfoundation/nodekit/api"
)

// InvalidStatus indicates an error when a response contains an invalid or unexpected status code.
const InvalidStatus = "invalid status"

// State represents the operational state of a process or system within the application. It is defined as a string type.
type State string

const (

	// FastCatchupState represents the state when the system is performing a fast catchup operation to synchronize.
	FastCatchupState State = "FAST-CATCHUP"

	// SyncingState represents the state where the system is in the process of synchronizing to the latest data or state.
	SyncingState State = "SYNCING"

	// StableState indicates the system is in a stable and operational state with no ongoing synchronization or major updates.
	StableState State = "RUNNING"
)

// Status represents the state of a system including metadata like version, network, and operational state.
type Status struct {

	// State represents the operational state of a process or system, defined as a string.
	State State `json:"state"`

	// Version represents the version identifier of the system, typically used to denote the current software version.
	Version string `json:"version"`

	// Network represents the name of the network the status is associated with.
	Network string `json:"network"`

	// Voting indicates whether a node participated in the current upgrade voting process.
	Voting bool `json:"voting"`

	// NeedsUpdate indicates whether the system requires an update based on the current version and available release data.
	NeedsUpdate bool `json:"needsUpdate"`

	// LastRound represents the most recent round number recorded by the system or client.
	LastRound uint64 `json:"lastRound"`

	// Catchpoint is a pointer to a string that identifies the current catchpoint for node synchronization or fast catchup.
	Catchpoint                  *string `json:"catchpoint"`
	CatchpointAccountsTotal     int     `json:"catchpointAccountsTotal"`
	CatchpointAccountsProcessed int     `json:"catchpointAccountsProcessed"`
	CatchpointAccountsVerified  int     `json:"catchpointAccountsVerified"`
	CatchpointKeyValueTotal     int     `json:"catchpointKeyValueTotal"`
	CatchpointKeyValueProcessed int     `json:"catchpointKeyValueProcessed"`
	CatchpointKeyValueVerified  int     `json:"catchpointKeyValueVerified"`
	CatchpointBlocksTotal       int     `json:"catchpointTotalBlocks"`
	CatchpointBlocksAcquired    int     `json:"catchpointAcquiredBlocks"`

	SyncTime int `json:"syncTime"`

	// Client provides methods for interacting with the API, adhering to ClientWithResponsesInterface specifications.
	Client api.ClientWithResponsesInterface `json:"-"`

	// HttpPkg represents an interface for HTTP package operations, providing methods for making HTTP requests.
	HttpPkg api.HttpPkgInterface `json:"-"`
}

// Update synchronizes non-identical fields between two Status instances and returns the updated Status.
func (s Status) Update(status Status) Status {
	if s.State != status.State {
		s.State = status.State
	}
	if s.Version != status.Version {
		s.Version = status.Version
	}
	if s.Network != status.Network {
		s.Network = status.Network
	}
	if s.Voting != status.Voting {
		s.Voting = status.Voting
	}
	if s.NeedsUpdate != status.NeedsUpdate {
		s.NeedsUpdate = status.NeedsUpdate
	}
	if s.LastRound != status.LastRound {
		s.LastRound = status.LastRound
	}
	return s
}

// Wait waits for the next block round based on the current LastRound and updates the Status with the returned response.
// It interacts with the client's WaitForBlockWithResponse method and handles any errors or invalid status codes.
// Returns the updated Status, the response object, or an error if the operation fails.
func (s Status) Wait(ctx context.Context) (Status, api.ResponseInterface, error) {
	response, err := s.Client.WaitForBlockWithResponse(ctx, int(s.LastRound))
	if err != nil {
		return s, response, err
	}
	if response.StatusCode() >= 300 {
		return s, response, errors.New(InvalidStatus)
	}

	return s.Merge(*response.JSON200), response, nil
}

// Merge updates the current Status with data from a given StatusLike instance and adjusts fields based on defined conditions.
func (s Status) Merge(res api.StatusLike) Status {
	s.LastRound = uint64(res.LastRound)
	catchpoint := res.Catchpoint
	if catchpoint != nil && *catchpoint != "" {
		s.State = FastCatchupState
		s.Catchpoint = catchpoint
		s.SyncTime = res.CatchupTime
		s.CatchpointAccountsTotal = *res.CatchpointTotalAccounts
		s.CatchpointAccountsProcessed = *res.CatchpointProcessedAccounts
		s.CatchpointAccountsVerified = *res.CatchpointVerifiedAccounts
		s.CatchpointKeyValueTotal = *res.CatchpointTotalKvs
		s.CatchpointKeyValueProcessed = *res.CatchpointProcessedKvs
		s.CatchpointKeyValueVerified = *res.CatchpointVerifiedKvs
		s.CatchpointBlocksAcquired = *res.CatchpointAcquiredBlocks
		s.CatchpointBlocksTotal = *res.CatchpointTotalBlocks
	} else if res.CatchupTime > 0 {
		s.SyncTime = res.CatchupTime
		s.State = SyncingState
	} else {
		s.State = StableState
	}

	if res.UpgradeNodeVote != nil {
		s.Voting = *res.UpgradeNodeVote
	}
	return s
}

// Get retrieves the current system status by invoking the client's GetStatusWithResponse method and merging the result.
// It returns the updated Status, the API response, or an error if the request fails or the status code is invalid.
func (s Status) Get(ctx context.Context) (Status, api.ResponseInterface, error) {
	statusResponse, err := s.Client.GetStatusWithResponse(ctx)
	if err != nil {
		return s, statusResponse, err
	}
	if statusResponse.StatusCode() >= 300 {
		return s, statusResponse, errors.New(InvalidStatus)
	}
	return s.Merge(*statusResponse.JSON200), statusResponse, nil
}

// NewStatus initializes and returns a Status object based on the provided context, client, and HTTP package interface.
// The function also checks for system updates and merges the current status with the latest available data.
func NewStatus(ctx context.Context, client api.ClientWithResponsesInterface, httpPkg api.HttpPkgInterface) (Status, api.ResponseInterface, error) {
	var status Status
	status.Client = client
	status.HttpPkg = httpPkg

	v, versionResponse, err := GetVersion(ctx, client)
	if err != nil {
		return status, versionResponse.(api.ResponseInterface), err
	}
	status.Network = v.Network
	status.Version = v.Version
	status.NeedsUpdate = false

	if v.Channel == "beta" || v.Channel == "stable" {
		// TODO: last checked
		releaseResponse, err := api.GetGoAlgorandReleaseWithResponse(httpPkg, v.Channel)
		// Return the error and response
		if err != nil {
			return status, releaseResponse, err
		}
		// Update status update field
		if releaseResponse != nil && status.Version != releaseResponse.JSON200 {
			status.NeedsUpdate = true
		}
	}

	return status.Get(ctx)
}
