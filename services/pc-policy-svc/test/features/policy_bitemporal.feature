Feature: Bi-temporal Policy Endorsement
  In order to maintain strict historical accuracy for insurance policies
  As a policy administrator
  I want to ensure that endorsements do not corrupt the policy timeline

  Scenario: Successive Endorsements
    Given an active policy exists with premium 1000.00
    When an endorsement is made effective "2026-08-01" with new premium 1200.00
    Then the policy should have 2 versions
    And the latest version should have premium 1200.00
    And the previous version should end effective "2026-08-01"

  Scenario: Backdated Endorsement Validation
    Given an active policy exists with effective date "2026-05-01"
    When an endorsement is made effective "2026-04-01"
    Then an invalid endorsement date error should be returned
