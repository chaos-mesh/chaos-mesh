Feature: NetworkChaos - Network Delay

  Background:
    Given the network peers are ready and all connections are good

  Scenario: Delay all outbound traffic from peer-0
    When I create a NetworkDelay chaos named "network-chaos-1" with latency "200ms" from "network-peer-0" in direction "to"
    Then slow connections should be peer pairs [[0,1],[0,2],[0,3]] within 15 seconds
    When I delete the NetworkChaos "network-chaos-1"
    Then all connections should recover within 15 seconds

  Scenario: Delay traffic from peer-0 to peer-1 only
    When I create a NetworkDelay chaos named "network-chaos-1" with latency "200ms" from "network-peer-0" to "network-peer-1" in direction "to"
    Then slow connections should be peer pairs [[0,1]] within 15 seconds
    When I delete the NetworkChaos "network-chaos-1"
    Then all connections should recover within 15 seconds

  Scenario: Delay from peer-0 to even partition
    When I create a NetworkDelay chaos named "network-chaos-2" with latency "200ms" from "network-peer-0" to partition "0" in direction "to"
    Then slow connections should be peer pairs [[0,2]] within 15 seconds
    When I delete the NetworkChaos "network-chaos-2"
    Then all connections should recover within 15 seconds

  Scenario: Both direction delay on peer-0 to even partition
    When I create a NetworkDelay chaos named "network-chaos-4" with latency "200ms" from "network-peer-0" to partition "0" in direction "both"
    Then slow connections should be peer pairs [[0,2],[2,0]] within 15 seconds bidirectional
    When I delete the NetworkChaos "network-chaos-4"
    Then all connections should recover within 15 seconds
