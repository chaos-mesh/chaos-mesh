Feature: NetworkChaos - Network Partition

  Background:
    Given the network peers are ready and all connections are good

  Scenario: Block outbound traffic from peer-0 to peer-1
    When I create a NetworkPartition chaos named "network-chaos-1" from "network-peer-0" to "network-peer-1" in direction "to"
    Then blocked connections should be peer pairs [[0,1]] within 15 seconds
    When I delete the NetworkChaos "network-chaos-1"
    Then all connections should recover within 15 seconds

  Scenario: Block both directions between peer-0 and peer-1
    When I create a NetworkPartition chaos named "network-chaos-1" from "network-peer-0" to "network-peer-1" in direction "both"
    Then blocked connections should be peer pairs [[0,1],[1,0]] within 15 seconds
    When I delete the NetworkChaos "network-chaos-1"
    Then all connections should recover within 15 seconds

  Scenario: Block inbound traffic to peer-0 from peer-1
    When I create a NetworkPartition chaos named "network-chaos-1" from "network-peer-0" to "network-peer-1" in direction "from"
    Then blocked connections should be peer pairs [[1,0]] within 15 seconds
    When I delete the NetworkChaos "network-chaos-1"
    Then all connections should recover within 15 seconds

  Scenario: Block peer-0 from odd partition in both directions
    When I create a NetworkPartition chaos named "network-chaos-1" from "network-peer-0" to partition "1" with all pods in direction "both"
    Then blocked connections should be peer pairs [[0,1],[1,0],[0,3],[3,0]] within 15 seconds
    When I delete the NetworkChaos "network-chaos-1"
    Then all connections should recover within 15 seconds

  Scenario: Block all outbound traffic from peer-0
    When I create a NetworkPartition chaos named "network-chaos-without-target" from "network-peer-0" with no target in direction "to"
    Then blocked connections should be peer pairs [[0,1],[0,2],[0,3]] within 15 seconds
    When I delete the NetworkChaos "network-chaos-without-target"
    Then all connections should recover within 15 seconds

  Scenario: Reject chaos injection on pods using hostNetwork
    Given a hostNetwork deployment named "network-peer-4" is running in the test namespace
    When I create a NetworkPartition chaos named "network-chaos-1" from "network-peer-4" to "network-peer-1" in direction "to"
    Then the chaos "network-chaos-1" should not inject into "network-peer-4" pods within 1 minute
