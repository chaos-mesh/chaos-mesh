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

@jvmchaos
Feature: JVMChaos Simulation

  Scenario: JVM latency slows down method calls
    Given a namespace is prepared
    And a java pod named "helloworld" printing one log line per second is running
    When a JVM latency chaos named "helloworld-latency" with 3000 ms latency is applied to class "Main" method "sayhello" for pods with label "app=helloworld"
    Then the pod named "helloworld" should eventually print log lines with intervals longer than 3000 milliseconds

  Scenario: JVM exception chaos injects and recovers
    Given a namespace is prepared
    And a java pod named "helloworld" printing one log line per second is running
    When a JVM exception chaos named "helloworld-exception" throwing exception "java.io.IOException" with message "BOOM" is applied to class "Main" method "sayhello" for pods with label "app=helloworld"
    Then the pod named "helloworld" should eventually print a log line containing "Got an exception! java.io.IOException: BOOM"
    When the JVM chaos "helloworld-exception" is deleted
    Then the pod named "helloworld" should eventually stop printing log lines containing "Got an exception! java.io.IOException: BOOM"

  Scenario: JVM ruleData chaos modifies return value and recovers
    Given a namespace is prepared
    And a java pod named "helloworld" printing one log line per second is running
    When a JVM ruleData chaos named "helloworld-rule" modifying method "getnum" of class "Main" to return "9999" is applied to pods with label "app=helloworld"
    Then the pod named "helloworld" should eventually print a log line containing "9999. Hello World"
    When the JVM chaos "helloworld-rule" is deleted
    Then the pod named "helloworld" should eventually stop printing log lines containing "9999. Hello World"
