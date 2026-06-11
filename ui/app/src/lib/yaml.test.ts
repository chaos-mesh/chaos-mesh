/*
 * Copyright 2026 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import yaml from './yaml'

describe('lib/yaml', () => {
  describe('dump', () => {
    it('should sort Kubernetes keys correctly', () => {
      const data = {
        spec: { selector: { namespaces: ['default'] } },
        metadata: {
          namespace: 'chaos-testing',
          name: 'burn-cpu',
          annotations: { foo: 'bar' },
          labels: { baz: 'qux' },
        },
        kind: 'StressChaos',
        apiVersion: 'chaos-mesh.org/v1alpha1',
      }

      const res = yaml.dump(data)

      const lines = res.split('\n').map((line) => line.trim())

      // Verify top-level order
      const apiVersionIdx = lines.findIndex((l) => l.startsWith('apiVersion:'))
      const kindIdx = lines.findIndex((l) => l.startsWith('kind:'))
      const metadataIdx = lines.findIndex((l) => l.startsWith('metadata:'))
      const specIdx = lines.findIndex((l) => l.startsWith('spec:'))

      expect(apiVersionIdx).toBeLessThan(kindIdx)
      expect(kindIdx).toBeLessThan(metadataIdx)
      expect(metadataIdx).toBeLessThan(specIdx)

      // Verify metadata order
      const metadataStart = lines.findIndex((l) => l.startsWith('metadata:'))
      const nameIdx = lines.findIndex((l) => l.startsWith('name: burn-cpu'))
      const namespaceIdx = lines.findIndex((l) => l.startsWith('namespace: chaos-testing'))
      const labelsIdx = lines.findIndex((l) => l.startsWith('labels:'))
      const annotationsIdx = lines.findIndex((l) => l.startsWith('annotations:'))

      expect(nameIdx).toBeGreaterThan(metadataStart)
      expect(namespaceIdx).toBeGreaterThan(nameIdx)
      expect(labelsIdx).toBeGreaterThan(namespaceIdx)
      expect(annotationsIdx).toBeGreaterThan(labelsIdx)
    })
  })
})
