package steps

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/medhen/pc-policy-svc/internal/domain/policy"
)

type policyTestState struct {
	pol *policy.Policy
	err error
}

func (s *policyTestState) anActivePolicyExistsWithPremium(premium float64) error {
	s.err = nil
	effFrom, _ := time.Parse("2006-01-02", "2026-01-01")
	effTo := effFrom.AddDate(1, 0, 0)
	p, err := policy.NewPolicy("tenant1", "MOT/001", uuid.New(), uuid.New(), []byte(`{}`), premium, effFrom, effTo)
	if err != nil {
		return err
	}
	p.Status = policy.StatusActive
	s.pol = p
	return nil
}

func (s *policyTestState) anActivePolicyExistsWithEffectiveDate(dateStr string) error {
	s.err = nil
	effFrom, _ := time.Parse("2006-01-02", dateStr)
	effTo := effFrom.AddDate(1, 0, 0)
	p, err := policy.NewPolicy("tenant1", "MOT/002", uuid.New(), uuid.New(), []byte(`{}`), 1000.0, effFrom, effTo)
	if err != nil {
		return err
	}
	p.Status = policy.StatusActive
	s.pol = p
	return nil
}

func (s *policyTestState) anEndorsementIsMadeEffectiveWithNewPremium(dateStr string, premium float64) error {
	effDate, _ := time.Parse("2006-01-02", dateStr)
	_, s.err = s.pol.Endorse(effDate, []byte(`{"endorsed":true}`), premium)
	return nil
}

func (s *policyTestState) anEndorsementIsMadeEffective(dateStr string) error {
	return s.anEndorsementIsMadeEffectiveWithNewPremium(dateStr, 1200.0)
}

func (s *policyTestState) thePolicyShouldHaveVersions(count int) error {
	if len(s.pol.Versions) != count {
		return fmt.Errorf("expected %d versions, got %d", count, len(s.pol.Versions))
	}
	return nil
}

func (s *policyTestState) theLatestVersionShouldHavePremium(premium float64) error {
	latest := s.pol.Versions[len(s.pol.Versions)-1]
	if latest.TotalPremium != premium {
		return fmt.Errorf("expected premium %f, got %f", premium, latest.TotalPremium)
	}
	return nil
}

func (s *policyTestState) thePreviousVersionShouldEndEffective(dateStr string) error {
	prev := s.pol.Versions[len(s.pol.Versions)-2]
	expectedDate, _ := time.Parse("2006-01-02", dateStr)
	if prev.EffectiveTo == nil {
		return fmt.Errorf("previous version EffectiveTo is nil")
	}
	if !prev.EffectiveTo.Equal(expectedDate) {
		return fmt.Errorf("expected EffectiveTo %v, got %v", expectedDate, *prev.EffectiveTo)
	}
	return nil
}

func (s *policyTestState) anInvalidEndorsementDateErrorShouldBeReturned() error {
	if s.err == nil {
		return fmt.Errorf("expected an error, got nil")
	}
	if !errors.Is(s.err, policy.ErrInvalidEndorsementDate) {
		return fmt.Errorf("expected ErrInvalidEndorsementDate, got %v", s.err)
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	state := &policyTestState{}

	ctx.Step(`^an active policy exists with premium (\d+\.\d+)$`, state.anActivePolicyExistsWithPremium)
	ctx.Step(`^an active policy exists with effective date "([^"]*)"$`, state.anActivePolicyExistsWithEffectiveDate)
	ctx.Step(`^an endorsement is made effective "([^"]*)" with new premium (\d+\.\d+)$`, state.anEndorsementIsMadeEffectiveWithNewPremium)
	ctx.Step(`^an endorsement is made effective "([^"]*)"$`, state.anEndorsementIsMadeEffective)
	ctx.Step(`^the policy should have (\d+) versions$`, state.thePolicyShouldHaveVersions)
	ctx.Step(`^the latest version should have premium (\d+\.\d+)$`, state.theLatestVersionShouldHavePremium)
	ctx.Step(`^the previous version should end effective "([^"]*)"$`, state.thePreviousVersionShouldEndEffective)
	ctx.Step(`^an invalid endorsement date error should be returned$`, state.anInvalidEndorsementDateErrorShouldBeReturned)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
