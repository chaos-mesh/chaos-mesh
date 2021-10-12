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

package clientpool

import (
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	pkgclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func TestClientPool(t *testing.T) {
	g := NewWithT(t)

	t.Run("client pool", func(t *testing.T) {
		defer mock.With("MockCreateK8sClient", func(config *rest.Config, options pkgclient.Options) (pkgclient.Client, error) {
			return nil, nil
		})()

		k8sClients, err := NewClientPool(&rest.Config{}, &runtime.Scheme{}, 5)
		g.Expect(err).ToNot(HaveOccurred())

		for i := 0; i < 6; i++ {
			_, err := k8sClients.Client(strconv.Itoa(i))
			g.Expect(err).ToNot(HaveOccurred())
		}

		// remain key 1, 2, 3, 4, 5 in cache, 0 is evicted
		g.Expect(k8sClients.Num()).To(Equal(5))

		_, err = k8sClients.Client("6")
		g.Expect(err).ToNot(HaveOccurred())

		// 6 in cache, and 1 is evict because it is the oldest key which is not used recently
		g.Expect(k8sClients.Contains("6")).To(Equal(true))
		g.Expect(k8sClients.Contains("1")).To(Equal(false))
	})
}
