package internal

import (
	"context"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/algod/participation"
	"time"

	"github.com/algorandfoundation/algorun-tui/api"
)

// StateModel represents the state of the application,
// including status, metrics, accounts, keys, and other configurations.
type StateModel struct {

	// Status represents the current status of the algod node,
	// including network state and round information.
	Status algod.Status

	// Metrics provides runtime statistics including
	// round time, transactions per second, and data transfer metrics.
	Metrics algod.Metrics

	// Accounts holds a mapping of account identifiers to their corresponding Account details.
	// This map is derived from the list of the type api.ParticipationKey
	Accounts map[string]Account

	// ParticipationKeys is a slice of participation keys used by the node
	// to interact with the blockchain and consensus protocol.
	ParticipationKeys *[]api.ParticipationKey

	// Admin indicates whether the current node has
	// admin privileges or capabilities enabled.
	Admin bool

	// Watching indicates whether the StateModel is actively monitoring
	// changes or processes in a background loop.
	// TODO: handle contexts instead of adding it to state (skill-issue zero)
	Watching bool

	// Client provides an interface for interacting with API endpoints,
	// enabling various node operations and data retrieval.
	Client api.ClientWithResponsesInterface
	// HttpPkg provides an interface for making HTTP requests,
	// enabling communication with external APIs or services.
	HttpPkg api.HttpPkgInterface

	// Context provides a context for managing cancellation,
	// deadlines, and request-scoped values in StateModel operations.
	// TODO: implement more of the context
	Context context.Context
}

func NewStateModel(ctx context.Context, client api.ClientWithResponsesInterface, httpPkg api.HttpPkgInterface) (*StateModel, api.ResponseInterface, error) {
	status, response, err := algod.NewStatus(ctx, client, httpPkg)
	if err != nil {
		return nil, response, err
	}
	metrics, response, err := algod.NewMetrics(ctx, client, httpPkg, status.LastRound)
	if err != nil {
		return nil, response, err
	}

	partKeys, partkeysResponse, err := participation.GetKeys(ctx, client)

	return &StateModel{
		Status:            status,
		Metrics:           metrics,
		Accounts:          ParticipationKeysToAccounts(partKeys),
		ParticipationKeys: partKeys,

		Admin:    true,
		Watching: true,

		Client:  client,
		HttpPkg: httpPkg,
		Context: ctx,
	}, partkeysResponse, nil
}

// TODO: handle in context loop
func (s *StateModel) waitAfterError(err error, cb func(model *StateModel, err error)) {
	if err != nil {
		s.Status.State = "DOWN"
		cb(nil, err)
		time.Sleep(time.Second * 3)
	}
}

// TODO: allow context to handle loop
func (s *StateModel) Watch(cb func(model *StateModel, err error), ctx context.Context, client api.ClientWithResponsesInterface) {
	var err error

	// Setup Defaults
	s.Watching = true
	if s.Metrics.Window == 0 {
		s.Metrics.Window = 100
	}

	// Fetch the latest Status
	s.Status, _, err = s.Status.Get(ctx)
	if err != nil {
		// callback immediately on error
		cb(nil, err)
	}

	// The main Loop
	// TODO: Refactor to Context
	for {
		if !s.Watching {
			break
		}
		// Abort on Fast-Catchup
		if s.Status.State == algod.FastCatchupState {
			time.Sleep(time.Second * 10)
			s.Status, _, err = s.Status.Get(ctx)
			if err != nil {
				cb(nil, err)
			}
			continue
		}

		// Wait for the next block
		s.Status, _, err = s.Status.Wait(ctx)
		s.waitAfterError(err, cb)
		if err != nil {
			continue
		}

		// Fetch Keys
		s.UpdateKeys()

		if s.Status.State == algod.SyncingState {
			cb(s, nil)
			continue
		}
		// Run Round Averages and RX/TX every 5 rounds
		if s.Status.LastRound%5 == 0 {
			s.Metrics, _, err = s.Metrics.Get(ctx, s.Status.LastRound)
			s.waitAfterError(err, cb)
			if err != nil {
				continue
			}
		}

		// Callback the current state to the app
		cb(s, nil)
	}
}

func (s *StateModel) Stop() {
	s.Watching = false
}

func (s *StateModel) UpdateAccounts() error {
	var err error
	s.Accounts, err = AccountsFromState(s, new(Clock), s.Client)
	return err
}

func (s *StateModel) UpdateKeys() {
	var err error
	s.ParticipationKeys, _, err = participation.GetKeys(s.Context, s.Client)
	if err != nil {
		s.Admin = false
	}
	if err == nil {
		s.Admin = true
		err = s.UpdateAccounts()
		if err != nil {
			// TODO: Handle error
		}
	}
}
