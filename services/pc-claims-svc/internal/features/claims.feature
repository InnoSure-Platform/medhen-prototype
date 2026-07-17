Feature: Settlement Saga Orchestration
  In order to ensure financial consistency across microservices
  As the Tier-0 Claims Management Engine
  I need to automatically deduct reserves and compensate failures when settlements are disbursed

  Scenario: Successful Multi-Payee Disbursement
    Given an approved settlement for "c-1234" with a gross loss of 50000 ETB
    And a policy deductible of 5000 ETB
    And a salvage value of 10000 ETB
    When the settlement saga executes
    Then the net settlement amount should be exactly 35000 ETB
    And the saga should emit 1 outbox event to "pc.billing.disbursement.command.v1"
    
  Scenario: Fractional Subrogation Netting
    Given an open claim "c-9999" with Reinsurer share set at 40 percent
    When a gross recovery of 10000 ETB is received
    Then the Reinsurer Share Base should be 4000 ETB
    And the EIC Retention Base should be 6000 ETB
