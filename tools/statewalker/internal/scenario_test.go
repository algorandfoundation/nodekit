package statewalker

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// scriptedScenario builds a two-phase scenario recording the call order and
// failing at the requested step.
func scriptedScenario(calls *[]string, failAt string) Scenario {
	phase := func(name string) Phase {
		return Phase{
			Name: name,
			Run: func(ctx context.Context, env *Env) error {
				*calls = append(*calls, name+":run")
				if failAt == name+":run" {
					return errors.New("boom")
				}
				return nil
			},
			Verify: func(ctx context.Context, env *Env) error {
				*calls = append(*calls, name+":verify")
				if failAt == name+":verify" {
					return errors.New("boom")
				}
				return nil
			},
		}
	}
	return Scenario{Name: "test", Phases: []Phase{phase("a"), phase("b")}}
}

func Test_Scenario_Execute_OrderAndTeardown(t *testing.T) {
	ctx := context.Background()
	var calls []string
	teardowns := 0
	scenario := scriptedScenario(&calls, "")

	err := scenario.Execute(ctx, &Env{}, func() error { teardowns++; return nil })
	if err != nil {
		t.Fatalf("expected success, got: %s", err)
	}
	want := "a:run,a:verify,b:run,b:verify"
	if got := strings.Join(calls, ","); got != want {
		t.Errorf("expected call order %s, got %s", want, got)
	}
	if teardowns != 1 {
		t.Errorf("expected teardown once on success, got %d", teardowns)
	}
}

func Test_Scenario_Execute_KeepOnSuccess(t *testing.T) {
	ctx := context.Background()
	var calls []string
	teardowns := 0
	scenario := scriptedScenario(&calls, "")
	scenario.KeepOnSuccess = true

	err := scenario.Execute(ctx, &Env{}, func() error { teardowns++; return nil })
	if err != nil {
		t.Fatalf("expected success, got: %s", err)
	}
	if teardowns != 0 {
		t.Errorf("expected no teardown with KeepOnSuccess, got %d", teardowns)
	}
}

func Test_Scenario_Execute_RunFailure(t *testing.T) {
	ctx := context.Background()
	var calls []string
	teardowns := 0
	scenario := scriptedScenario(&calls, "a:run")

	err := scenario.Execute(ctx, &Env{}, func() error { teardowns++; return nil })
	if err == nil {
		t.Fatal("expected the scenario to fail")
	}
	if !strings.Contains(err.Error(), "phase a (run)") {
		t.Errorf("expected the error to name the failing phase, got: %s", err)
	}
	want := "a:run"
	if got := strings.Join(calls, ","); got != want {
		t.Errorf("expected execution to stop at the failure, got %s", got)
	}
	if teardowns != 1 {
		t.Errorf("expected teardown once on failure, got %d", teardowns)
	}
}

func Test_Scenario_Execute_VerifyFailure(t *testing.T) {
	ctx := context.Background()
	var calls []string
	teardowns := 0
	scenario := scriptedScenario(&calls, "b:verify")

	err := scenario.Execute(ctx, &Env{}, func() error { teardowns++; return nil })
	if err == nil {
		t.Fatal("expected the scenario to fail")
	}
	if !strings.Contains(err.Error(), "phase b (verify)") {
		t.Errorf("expected the error to name the failing verify, got: %s", err)
	}
	if teardowns != 1 {
		t.Errorf("expected teardown once on failure, got %d", teardowns)
	}
}

func Test_Scenario_Execute_KeepSkipsTeardown(t *testing.T) {
	ctx := context.Background()
	var calls []string
	teardowns := 0
	scenario := scriptedScenario(&calls, "a:run")

	err := scenario.Execute(ctx, &Env{Keep: true}, func() error { teardowns++; return nil })
	if err == nil {
		t.Fatal("expected the scenario to fail")
	}
	if teardowns != 0 {
		t.Errorf("expected no teardown with Keep, got %d", teardowns)
	}

	// Keep also skips the teardown on success
	calls = nil
	scenario = scriptedScenario(&calls, "")
	err = scenario.Execute(ctx, &Env{Keep: true}, func() error { teardowns++; return nil })
	if err != nil {
		t.Fatalf("expected success, got: %s", err)
	}
	if teardowns != 0 {
		t.Errorf("expected no teardown with Keep on success, got %d", teardowns)
	}
}
