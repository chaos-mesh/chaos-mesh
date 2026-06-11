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
    it('should sort Kubernetes keys correctly and preserve nested order', () => {
      const data = {
        spec: {
          z: 'z-val',
          a: 'a-val',
        },
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

      // Verify metadata block relative ordering using substring indices
      const metadataIdx = res.indexOf('metadata:')
      const nameIdx = res.indexOf('name: burn-cpu')
      const namespaceIdx = res.indexOf('namespace: chaos-testing')
      const labelsIdx = res.indexOf('labels:')
      const annotationsIdx = res.indexOf('annotations:')

      expect(metadataIdx).not.toBe(-1)
      expect(nameIdx).toBeGreaterThan(metadataIdx)
      expect(namespaceIdx).toBeGreaterThan(nameIdx)
      expect(labelsIdx).toBeGreaterThan(namespaceIdx)
      expect(annotationsIdx).toBeGreaterThan(labelsIdx)

      // Verify that nested keys under spec preserve their original order
      const specIdx = res.indexOf('spec:')
      const zIdx = res.indexOf('z: z-val')
      const aIdx = res.indexOf('a: a-val')

      expect(specIdx).not.toBe(-1)
      expect(zIdx).toBeGreaterThan(specIdx)
      expect(aIdx).toBeGreaterThan(zIdx) // z was inserted first, so it should stay first
    })
  })
})
