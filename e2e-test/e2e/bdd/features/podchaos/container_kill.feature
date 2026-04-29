Feature: PodChaos - Container Kill

  Scenario: Container kill once then delete
    Given a nginx deployment named "nginx" with 1 replica is running in the test namespace
    When I create a ContainerKill chaos named "nginx-container-kill" targeting container "nginx" in pods with label "app=nginx"
    Then the nginx container should show a last termination state within 5 minutes
    When I delete the PodChaos "nginx-container-kill"
    Then the nginx container should recover to running state within 5 minutes
