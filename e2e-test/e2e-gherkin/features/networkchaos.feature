# Copyright 2026 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

@networkchaos
Feature: NetworkChaos Simulation

  Scenario: NetworkDelay without target
    Given a namespace is prepared
    And 4 network peers are running
    When a "NetworkDelay" chaos named "network-delay" is applied on pods with label "app=network-peer-0"
    Then the network condition between peers should eventually be slow
    When the chaos experiment "network-delay" is deleted
    Then the network condition between peers should eventually recover

  Scenario: NetworkDelay with target
    Given a namespace is prepared
    And 4 network peers are running
    When a "NetworkDelay" chaos named "network-delay" with target "app=network-peer-1" is applied on pods with label "app=network-peer-0"
    Then the network condition from "network-peer-0" to "network-peer-1" should eventually be slow
    When the chaos experiment "network-delay" is deleted
    Then the network condition between peers should eventually recover

  Scenario: NetworkDelay with target partition
    Given a namespace is prepared
    And 4 network peers are running
    When a "NetworkDelay" chaos named "network-delay" with target partition "partition=0" is applied on pods with label "app=network-peer-0"
    Then the network condition from "network-peer-0" to partition "0" should eventually be slow
    When the chaos experiment "network-delay" is deleted
    Then the network condition between peers should eventually recover

  Scenario: NetworkDelay with both directions
    Given a namespace is prepared
    And 4 network peers are running
    When a "NetworkDelay" chaos named "network-delay" with target partition "partition=0" in both directions is applied on pods with label "app=network-peer-0"
    Then the network condition in both directions between "network-peer-0" and partition "0" should eventually be slow
    When the chaos experiment "network-delay" is deleted
    Then the network condition between peers should eventually recover

  Scenario: NetworkDelay with single external target
    Given a namespace is prepared
    And 4 network peers are running
    When a "NetworkDelay" chaos named "external-delay" with single external target "network-peer-1" is applied on pods with label "app=network-peer-0"
    Then the network condition from "network-peer-0" to "network-peer-1" should eventually be slow
    When the chaos experiment "external-delay" is deleted
    Then the network condition between peers should eventually recover

  Scenario: NetworkDelay with multiple external targets
    Given a namespace is prepared
    And 4 network peers are running
    When a "NetworkDelay" chaos named "external-delay" with external targets "network-peer-1" and "network-peer-2" is applied on pods with label "app=network-peer-0"
    Then the network condition from "network-peer-0" to "network-peer-1" and "network-peer-2" should eventually be slow
    When the chaos experiment "external-delay" is deleted
    Then the network condition between peers should eventually recover
