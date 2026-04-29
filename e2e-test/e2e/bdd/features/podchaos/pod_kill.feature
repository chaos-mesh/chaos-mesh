Feature: PodChaos - Pod Kill

  Scenario: Pod kill once then delete
    Given a nginx pod named "nginx" is running in the test namespace
    When I create a PodKill chaos named "nginx-kill" targeting pods with label "app=nginx"
    Then the pod "nginx" should be deleted within 5 minutes
    When I delete the PodChaos "nginx-kill"

  Scenario: Pod kill pause then unpause
    Given a nginx deployment named "nginx" with 3 replicas is running in the test namespace
    When I create a PodKill chaos named "nginx-kill" targeting pods with label "app=nginx"
    Then at least one pod UID should change within 5 minutes
    When I pause the PodChaos "nginx-kill"
    Then the chaos "nginx-kill" should NOT enter stopped phase within 5 seconds
    And no pod UIDs should change within 1 minute
