// Copyright 2026 Chaos Mesh Authors.
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

package tproxyconfig

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTLSConfigNumericContentsAcrossDaemon(t *testing.T) {
	for _, withCA := range []bool{false, true} {
		name := "without CA"
		if withCA {
			name = "with CA"
		}
		t.Run(name, func(t *testing.T) {
			controllerTLS := TLSConfig{
				CertFile: TLSConfigItem{Type: "Contents", Value: []byte{0, 45, 127, 128, 255}},
				KeyFile:  TLSConfigItem{Type: "Contents", Value: []byte{255, 128, 127, 45, 0}},
			}
			if withCA {
				controllerTLS.CAFile = &TLSConfigItem{Type: "Contents", Value: []byte{45, 0, 255, 127, 128}}
			}
			requestTLS, err := json.Marshal(controllerTLS)
			require.NoError(t, err)
			assertNumericTLS(t, requestTLS, withCA)

			var daemonTLS TLSConfig
			require.NoError(t, json.Unmarshal(requestTLS, &daemonTLS))
			configJSON, err := json.Marshal(Config{TLS: &daemonTLS})
			require.NoError(t, err)
			var config map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(configJSON, &config))
			assertNumericTLS(t, config["tls"], withCA)
		})
	}
}

func TestTLSConfigLegacyBase64InputBecomesNumericContents(t *testing.T) {
	const legacy = `{"cert_file":{"type":"Contents","value":"AC1/gP8="},"key_file":{"type":"Contents","value":"/4B/LQA="},"ca_file":{"type":"Contents","value":"LQD/f4A="}}`
	var daemonTLS TLSConfig
	require.NoError(t, json.Unmarshal([]byte(legacy), &daemonTLS))
	configJSON, err := json.Marshal(Config{TLS: &daemonTLS})
	require.NoError(t, err)
	var config map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(configJSON, &config))
	assertNumericTLS(t, config["tls"], true)
}

func TestTLSConfigEmptyContentsAreArrays(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value []byte
	}{
		{name: "nil"},
		{name: "empty", value: []byte{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := TLSConfigItem{Type: "Contents", Value: tc.value}
			requestTLS, err := json.Marshal(TLSConfig{CertFile: item, KeyFile: item, CAFile: &item})
			require.NoError(t, err)
			const expected = `{"cert_file":{"type":"Contents","value":[]},"key_file":{"type":"Contents","value":[]},"ca_file":{"type":"Contents","value":[]}}`
			require.JSONEq(t, expected, string(requestTLS))
			var daemonTLS TLSConfig
			require.NoError(t, json.Unmarshal(requestTLS, &daemonTLS))
			configJSON, err := json.Marshal(Config{TLS: &daemonTLS})
			require.NoError(t, err)
			var config map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(configJSON, &config))
			require.JSONEq(t, expected, string(config["tls"]))
		})
	}
}

func TestConfigOmitsAbsentTLS(t *testing.T) {
	configJSON, err := json.Marshal(Config{})
	require.NoError(t, err)
	require.JSONEq(t, `{"rules":null}`, string(configJSON))
}

func assertNumericTLS(t *testing.T, data []byte, withCA bool) {
	t.Helper()
	var tls map[string]struct {
		Type  string `json:"type"`
		Value []int  `json:"value"`
	}
	require.NoError(t, json.Unmarshal(data, &tls))
	expected := map[string][]int{
		"cert_file": {0, 45, 127, 128, 255},
		"key_file":  {255, 128, 127, 45, 0},
	}
	if withCA {
		expected["ca_file"] = []int{45, 0, 255, 127, 128}
	}
	require.Len(t, tls, len(expected))
	for field, values := range expected {
		item, ok := tls[field]
		require.True(t, ok, field)
		require.Equal(t, "Contents", item.Type, field)
		require.Equal(t, values, item.Value, field)
	}
}
