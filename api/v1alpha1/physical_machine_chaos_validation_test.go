// Copyright 2021 Chaos Mesh Authors.
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

package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_validateRedisExpirationAction(t *testing.T) {
	g := NewWithT(t)

	t.Run("valid option XX", func(t *testing.T) {
		err := validateRedisExpirationAction(&RedisExpirationSpec{
			RedisCommonSpec: RedisCommonSpec{Addr: "localhost:6379"},
			Option:          "XX",
		})
		g.Expect(err).To(BeNil())
	})

	t.Run("valid option NX", func(t *testing.T) {
		err := validateRedisExpirationAction(&RedisExpirationSpec{
			RedisCommonSpec: RedisCommonSpec{Addr: "localhost:6379"},
			Option:          "NX",
		})
		g.Expect(err).To(BeNil())
	})

	t.Run("invalid option", func(t *testing.T) {
		err := validateRedisExpirationAction(&RedisExpirationSpec{
			RedisCommonSpec: RedisCommonSpec{Addr: "localhost:6379"},
			Option:          "INVALID",
		})
		g.Expect(err).NotTo(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("option invalid"))
	})

	t.Run("empty option is valid (no option specified)", func(t *testing.T) {
		err := validateRedisExpirationAction(&RedisExpirationSpec{
			RedisCommonSpec: RedisCommonSpec{Addr: "localhost:6379"},
			Option:          "",
		})
		g.Expect(err).To(BeNil())
	})

	t.Run("valid option GT", func(t *testing.T) {
		err := validateRedisExpirationAction(&RedisExpirationSpec{
			RedisCommonSpec: RedisCommonSpec{Addr: "localhost:6379"},
			Option:          "GT",
		})
		g.Expect(err).To(BeNil())
	})

	t.Run("valid option LT", func(t *testing.T) {
		err := validateRedisExpirationAction(&RedisExpirationSpec{
			RedisCommonSpec: RedisCommonSpec{Addr: "localhost:6379"},
			Option:          "LT",
		})
		g.Expect(err).To(BeNil())
	})

	t.Run("missing addr returns error", func(t *testing.T) {
		err := validateRedisExpirationAction(&RedisExpirationSpec{
			Option: "XX",
		})
		g.Expect(err).NotTo(BeNil())
		g.Expect(err.Error()).To(ContainSubstring("addr"))
	})
}
