---
name: BDD Probe request
about: Suggest new BDD specification
title: "[BDD Probe]"
labels: feature request, bdd, probe
assignees: ''

---
**Please write BDD specification**
_[Replace sample below with actual BDD specification. See Gherkin syntax for reference: https://cucumber.io/docs/gherkin/reference]_
```
Feature: Guess the word

  # The first example has two steps
  Scenario: Maker starts a game
    When the Maker starts a game
    Then the Maker waits for a Breaker to join
```

**Who is the SME validating this scenario?**
_[Enter name of SME(s) here]_

**For developer: Please describe how you are planning to implement the scenario above**
_[Assigned developer shall use this section to describe planned implementation. This shall be validated with SME before code is written]
[Replace sample content with actual steps]_
| Scenario Step | Implementation Plan |
|---|---|
|When the Maker starts a game|Call api endpoint and start game service|
|Then the Maker waits for a Breaker to join|Call api end point and check status is Waiting|

**For SME: Is the above implementation plan correct?**
_[SME shall validate above implementation plan, providing green light for implementation]_