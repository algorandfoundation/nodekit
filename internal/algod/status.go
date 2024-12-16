package algod

import (
	"context"
	"errors"
	"github.com/algorandfoundation/algorun-tui/api"
)

type State string

const (
	FastCatchupState State = "FAST-CATCHUP"
	SyncingState     State = "SYNCING"
	StableState      State = "RUNNING"
)

// Status represents a status response from algod.Status
type Status struct {
	State       State
	Version     string
	Network     string
	Voting      bool
	NeedsUpdate bool
	LastRound   uint64 // Last recorded round

	Client  api.ClientWithResponsesInterface
	HttpPkg api.HttpPkgInterface
}

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

func (s Status) WaitForStatus(ctx context.Context) (Status, api.ResponseInterface, error) {
	response, err := s.Client.WaitForBlockWithResponse(ctx, int(s.LastRound))
	if err != nil {
		return s, response, err
	}
	if response.StatusCode() >= 300 {
		return s, response, errors.New("status error")
	}

	return s.Merge(*response.JSON200), response, nil
}
func (s Status) Merge(res api.StatusLike) Status {
	s.LastRound = uint64(res.LastRound)
	catchpoint := res.Catchpoint
	if catchpoint != nil && *catchpoint != "" {
		s.State = FastCatchupState
	} else if res.CatchupTime > 0 {
		s.State = SyncingState
	} else {
		s.State = StableState
	}

	if res.UpgradeNodeVote != nil {
		s.Voting = *res.UpgradeNodeVote
	}
	return s
}
func (s Status) Get(ctx context.Context) (Status, api.ResponseInterface, error) {
	statusResponse, err := s.Client.GetStatusWithResponse(ctx)
	if err != nil {
		return s, statusResponse, err
	}
	if statusResponse.StatusCode() >= 300 {
		return s, statusResponse, errors.New("status error")
	}
	return s.Merge(*statusResponse.JSON200), statusResponse, nil
}
func NewStatus(ctx context.Context, client api.ClientWithResponsesInterface, httpPkg api.HttpPkgInterface) (Status, api.ResponseInterface, error) {
	var status Status
	status.Client = client
	status.HttpPkg = httpPkg

	v, versionResponse, err := GetVersion(ctx, client)
	if err != nil {
		return status, versionResponse, err
	}
	status.Network = v.Network
	status.Version = v.Version

	// TODO: last checked
	releaseResponse, err := api.GetGoAlgorandReleaseWithResponse(httpPkg, v.Channel)
	// Return the error and response
	if err != nil {
		return status, releaseResponse, err
	}
	// Update status update field
	if releaseResponse != nil && status.Version != releaseResponse.JSON200 {
		status.NeedsUpdate = true
	} else {
		status.NeedsUpdate = false
	}

	return status.Get(ctx)
}
