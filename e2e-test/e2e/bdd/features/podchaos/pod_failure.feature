Feature: PodChaos - Pod Failure

  Scenario: Pod failure once then delete
    Given a timer deployment named "timer-pod-failure1" is running in the test namespace
    When I create a PodFailure chaos named "timer-failure1" targeting pods with label "app=timer-pod-failure1"
    Then a pod in deployment "timer-pod-failure1" should have its container image replaced with the pause image
    When I delete the PodChaos "timer-failure1"
    Then all pods in deployment "timer-pod-failure1" should recover to the original image within 2 minutes

  Scenario: Pod failure pause then unpause
    Given a timer deployment named "timer-pod-failure2" is running in the test namespace
    When I create a PodFailure chaos with duration "9m" named "timer-failure2" targeting pods with label "app=timer-pod-failure2"
    Then a pod in deployment "timer-pod-failure2" should have its container image replaced with the pause image
    When I pause the PodChaos "timer-failure2"
    Then the chaos "timer-failure2" should enter stopped phase within 5 minutes
    And no pod in deployment "timer-pod-failure2" should have the pause image within 30 seconds
    When I unpause the PodChaos "timer-failure2"
    Then the chaos "timer-failure2" should enter running phase within 5 minutes
    And a pod in deployment "timer-pod-failure2" should have its container image replaced with the pause image again
