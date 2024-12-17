package internal

import (
	"context"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/algod/participation"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/charmbracelet/log"
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
	Accounts map[string]algod.Account

	// ParticipationKeys is a slice of participation keys used by the node
	// to interact with the blockchain and consensus protocol.
	ParticipationKeys participation.List

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

// NewStateModel initializes and returns a new StateModel instance
// along with an API response and potential error.
func NewStateModel(ctx context.Context, client api.ClientWithResponsesInterface, httpPkg api.HttpPkgInterface) (*StateModel, api.ResponseInterface, error) {
	// Preload the node status
	status, response, err := algod.NewStatus(ctx, client, httpPkg)
	if response == nil {
		log.Fatal("woooop")
	}
	if err != nil {
		return nil, response, err
	}
	// Try to fetch the latest metrics
	metrics, response, err := algod.NewMetrics(ctx, client, httpPkg, status.LastRound)
	if err != nil {
		return nil, response, err
	}

	partKeys, partkeysResponse, err := participation.GetList(ctx, client)

	return &StateModel{
		Status:            status,
		Metrics:           metrics,
		Accounts:          algod.ParticipationKeysToAccounts(partKeys),
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
func (s *StateModel) Watch(cb func(model *StateModel, err error), ctx context.Context, t system.Time) {
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
		s.UpdateKeys(t)

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

func (s *StateModel) UpdateKeys(t system.Time) {
	var err error
	s.ParticipationKeys, _, err = participation.GetList(s.Context, s.Client)
	if err != nil {
		s.Admin = false
	}
	if err == nil {
		s.Admin = true
		s.Accounts = algod.ParticipationKeysToAccounts(s.ParticipationKeys)
		for _, acct := range s.Accounts {
			// For each account, update the data from the RPC endpoint
			if s.Status.State == algod.StableState {
				// Skip eon errors
				rpcAcct, err := algod.GetAccount(s.Client, acct.Address)
				if err != nil {
					continue
				}

				s.Accounts[acct.Address] = s.Accounts[acct.Address].Merge(rpcAcct)
				s.Accounts[acct.Address] = s.Accounts[acct.Address].UpdateExpiredTime(t, s.ParticipationKeys, int(s.Status.LastRound), s.Metrics.RoundTime)
			}
		}
	}
}
