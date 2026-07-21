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
