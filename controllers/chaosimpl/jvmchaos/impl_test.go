// Copyright 2022 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package jvmchaos

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestGenerateRuleData(t *testing.T) {
	g := NewWithT(t)

	testCases := []struct {
		spec     *v1alpha1.JVMChaosSpec
		ruleData string
	}{
		{
			&v1alpha1.JVMChaosSpec{
				Action: v1alpha1.JVMExceptionAction,
				JVMParameter: v1alpha1.JVMParameter{
					Name: "test",
					JVMCommonSpec: v1alpha1.JVMCommonSpec{
						Pid: 1234,
					},
					JVMClassMethodSpec: v1alpha1.JVMClassMethodSpec{
						Class:  "testClass",
						Method: "testMethod",
					},
					ThrowException: "java.io.IOException(\"BOOM\")",
				},
			},
			"\nRULE test\nCLASS testClass\nMETHOD testMethod\nAT ENTRY\nIF true\nDO\n\tthrow new java.io.IOException(\"BOOM\");\nENDRULE\n",
		},
		{
			&v1alpha1.JVMChaosSpec{
				Action: v1alpha1.JVMReturnAction,
				JVMParameter: v1alpha1.JVMParameter{
					Name: "test",
					JVMCommonSpec: v1alpha1.JVMCommonSpec{
						Pid: 1234,
					},
					JVMClassMethodSpec: v1alpha1.JVMClassMethodSpec{
						Class:  "testClass",
						Method: "testMethod",
					},
					ReturnValue: "\"test\"",
				},
			},
			"\nRULE test\nCLASS testClass\nMETHOD testMethod\nAT ENTRY\nIF true\nDO\n\treturn \"test\";\nENDRULE\n",
		},
		{
			&v1alpha1.JVMChaosSpec{
				Action: v1alpha1.JVMLatencyAction,
				JVMParameter: v1alpha1.JVMParameter{
					Name: "test",
					JVMCommonSpec: v1alpha1.JVMCommonSpec{
						Pid: 1234,
					},
					JVMClassMethodSpec: v1alpha1.JVMClassMethodSpec{
						Class:  "testClass",
						Method: "testMethod",
					},
					LatencyDuration: 5000,
				},
			},
			"\nRULE test\nCLASS testClass\nMETHOD testMethod\nAT ENTRY\nIF true\nDO\n\tThread.sleep(5000);\nENDRULE\n",
		},
		{
			&v1alpha1.JVMChaosSpec{
				Action: v1alpha1.JVMStressAction,
				JVMParameter: v1alpha1.JVMParameter{
					Name: "test",
					JVMCommonSpec: v1alpha1.JVMCommonSpec{
						Pid: 1234,
					},
					JVMStressCfgSpec: v1alpha1.JVMStressCfgSpec{
						CPUCount: 1,
					},
				},
			},
			"\nRULE test\nCLASS org.chaos_mesh.chaos_agent.TriggerThread\nMETHOD triggerFunc\nHELPER org.chaos_mesh.byteman.helper.StressHelper\nAT ENTRY\nBIND flag:boolean=true;\nIF true\nDO\n\tinjectCPUStress(\"test\", 1);\nENDRULE\n",
		},
		{
			&v1alpha1.JVMChaosSpec{
				Action: v1alpha1.JVMStressAction,
				JVMParameter: v1alpha1.JVMParameter{
					Name: "test",
					JVMCommonSpec: v1alpha1.JVMCommonSpec{
						Pid: 1234,
					},
					JVMStressCfgSpec: v1alpha1.JVMStressCfgSpec{
						MemoryType: "heap",
					},
				},
			},
			"\nRULE test\nCLASS org.chaos_mesh.chaos_agent.TriggerThread\nMETHOD triggerFunc\nHELPER org.chaos_mesh.byteman.helper.StressHelper\nAT ENTRY\nBIND flag:boolean=true;\nIF true\nDO\n\tinjectMemStress(\"test\", \"heap\");\nENDRULE\n",
		},
		{
			&v1alpha1.JVMChaosSpec{
				Action: v1alpha1.JVMGCAction,
				JVMParameter: v1alpha1.JVMParameter{
					Name: "test",
					JVMCommonSpec: v1alpha1.JVMCommonSpec{
						Pid: 1234,
					},
				},
			},
			"\nRULE test\nCLASS org.chaos_mesh.chaos_agent.TriggerThread\nMETHOD triggerFunc\nHELPER org.chaos_mesh.byteman.helper.GCHelper\nAT ENTRY\nBIND flag:boolean=true;\nIF true\nDO\n\tgc();\nENDRULE\n",
		},
		{
			&v1alpha1.JVMChaosSpec{
				Action: v1alpha1.JVMMySQLAction,
				JVMParameter: v1alpha1.JVMParameter{
					Name: "test",
					JVMCommonSpec: v1alpha1.JVMCommonSpec{
						Pid: 1234,
					},
					JVMMySQLSpec: v1alpha1.JVMMySQLSpec{
						MySQLConnectorVersion: "8",
						Database:              "test",
						Table:                 "t1",
						SQLType:               "select",
					},
					ThrowException: "BOOM",
				},
			},
			"\nRULE test\nCLASS com.mysql.cj.NativeSession\nMETHOD execSQL\nHELPER org.chaos_mesh.byteman.helper.SQLHelper\nAT ENTRY\nBIND flag:boolean=matchDBTable(\"\", $2, \"test\", \"t1\", \"select\");\nIF flag\nDO\n\tthrow new com.mysql.cj.exceptions.CJException(\"BOOM\");\nENDRULE\n",
		},
		{
			&v1alpha1.JVMChaosSpec{
				Action: v1alpha1.JVMMySQLAction,
				JVMParameter: v1alpha1.JVMParameter{
					Name: "test",
					JVMCommonSpec: v1alpha1.JVMCommonSpec{
						Pid: 1234,
					},
					JVMMySQLSpec: v1alpha1.JVMMySQLSpec{
						MySQLConnectorVersion: "8",
						Database:              "test",
						Table:                 "t1",
						SQLType:               "select",
					},
					LatencyDuration: 5000,
				},
			},
			"\nRULE test\nCLASS com.mysql.cj.NativeSession\nMETHOD execSQL\nHELPER org.chaos_mesh.byteman.helper.SQLHelper\nAT ENTRY\nBIND flag:boolean=matchDBTable(\"\", $2, \"test\", \"t1\", \"select\");\nIF flag\nDO\n\tThread.sleep(5000);\nENDRULE\n",
		},
	}

	for _, testCase := range testCases {
		err := generateRuleData(testCase.spec)
		g.Expect(err).ShouldNot(HaveOccurred())
		g.Expect(testCase.spec.RuleData).Should(Equal(testCase.ruleData))
	}
}
