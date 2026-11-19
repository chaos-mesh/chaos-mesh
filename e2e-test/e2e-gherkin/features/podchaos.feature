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

@podchaos
Feature: PodChaos Simulation

  Scenario: PodKill once then delete
    Given a namespace is prepared
    And a single pod named "nginx" is running
    When a "PodKill" chaos named "nginx-kill" with mode "one" is applied to pods with label "app=nginx"
    Then the pod named "nginx" should eventually not be found

  Scenario: PodKill pause does not trigger further kills
    Given a namespace is prepared
    And a deployment named "nginx" with 3 replicas is running
    When the initial pod UIDs are recorded
    And a "PodKill" chaos named "nginx-kill" with mode "one" is applied to pods with label "app=nginx"
    Then at least one pod should be replaced with a different UID
    When the chaos experiment "nginx-kill" is paused
    Then no further pods should be killed within 1 minute

  Scenario: PodFailure once then delete
    Given a namespace is prepared
    And a deployment named "nginx" with 1 replicas is running
    When a "PodFailure" chaos named "nginx-failure" with mode "one" is applied to pods with label "app=nginx"
    Then the pods with label "app=nginx" should eventually have their container image changed to the pause image
    When the chaos experiment "nginx-failure" is deleted
    Then the pods with label "app=nginx" should eventually recover their original container image

  Scenario: PodFailure pause then unpause
    Given a namespace is prepared
    And a deployment named "nginx" with 1 replicas is running
    When a "PodFailure" chaos named "nginx-failure" with mode "one" is applied to pods with label "app=nginx"
    Then the pods with label "app=nginx" should eventually have their container image changed to the pause image
    When the chaos experiment "nginx-failure" is paused
    Then the pods with label "app=nginx" should eventually recover their original container image
    When the chaos experiment "nginx-failure" is unpaused
    Then the pods with label "app=nginx" should eventually have their container image changed to the pause image

  Scenario: ContainerKill once then delete
    Given a namespace is prepared
    And a deployment named "nginx" with 1 replicas is running
    When a "ContainerKill" chaos named "nginx-container-kill" with mode "one" targeting container "nginx" is applied to pods with label "app=nginx"
    Then the container "nginx" in pods with label "app=nginx" should eventually be terminated
    When the chaos experiment "nginx-container-kill" is deleted
    Then the container "nginx" in pods with label "app=nginx" should eventually be running and ready

  Scenario: ContainerKill pause then unpause
    Given a namespace is prepared
    And a deployment named "nginx" with 1 replicas is running
    When the container ID of container "nginx" in pods with label "app=nginx" is recorded
    And a "ContainerKill" chaos named "nginx-container-kill" with mode "one" targeting container "nginx" is applied to pods with label "app=nginx"
    Then the container ID should change
    When the chaos experiment "nginx-container-kill" is paused
    And the container ID of container "nginx" in pods with label "app=nginx" is recorded again
    Then the container ID should not change within 1 minute
    When the chaos experiment "nginx-container-kill" is unpaused
    Then the container ID should not change within 10 seconds
