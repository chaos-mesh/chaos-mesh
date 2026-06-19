Feature: PodChaos Simulation

  Scenario: PodKill once then delete
    Given a namespace is prepared
    And a single pod named "nginx" is running
    When a "PodKill" chaos named "nginx-kill" is applied to pods with label "app=nginx"
    Then the pod named "nginx" should eventually not be found

  Scenario: PodKill pause does not trigger further kills
    Given a namespace is prepared
    And a deployment named "nginx" with 3 replicas is running
    When the initial pod UIDs are recorded
    And a "PodKill" chaos named "nginx-kill" is applied to pods with label "app=nginx"
    Then at least one pod should be replaced with a different UID
    When the chaos experiment "nginx-kill" is paused
    Then no further pods should be killed within 1 minute
