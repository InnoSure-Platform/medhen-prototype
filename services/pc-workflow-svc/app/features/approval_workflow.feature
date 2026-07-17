Feature: Workflow SLA and Escaltion
  As an NBE auditor
  I want to ensure tasks that breach their SLA are escalated to the branch manager
  So that approval chains don't stall

  Scenario: A task breaches SLA and is escalated
    Given the workflow service is running
    When an approval workflow is initiated for quote "QT-100"
    And the SLA timer expires
    Then the task should be reassigned to the "BranchManager"
    And an "ESCALATED" entry should be written to the approval history
