Feature: Product Lifecycle Management
  In order to manage insurance products safely
  As a Product Manager
  I want to create and version products with strict state constraints

  Scenario: Create a new Product Draft
    Given an authenticated Product Manager
    When they submit a CreateProduct command for "MOT-01"
    Then the product is persisted in the "DRAFT" state
    And a ProductDraftCreated event is published to "platform.product.lifecycle.v1"
