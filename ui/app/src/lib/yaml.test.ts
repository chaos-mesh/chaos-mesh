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
    it('should sort Kubernetes keys correctly, handle embedded resources, and preserve nested order', () => {
      const data = {
        spec: {
          z: 'z-val',
          a: 'a-val',
          embeddedResource: {
            spec: {
              x: 'x-val',
              b: 'b-val',
            },
            metadata: {
              name: 'embedded-name',
            },
            kind: 'Pod',
            apiVersion: 'v1',
          },
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

      // 1. Verify root-level canonical ordering using substring indices
      const rootApiVersionIdx = res.indexOf('apiVersion: chaos-mesh.org/v1alpha1')
      const rootKindIdx = res.indexOf('kind: StressChaos')
      const rootMetadataIdx = res.indexOf('metadata:')
      const rootSpecIdx = res.indexOf('spec:')

      expect(rootApiVersionIdx).not.toBe(-1)
      expect(rootKindIdx).toBeGreaterThan(rootApiVersionIdx)
      expect(rootMetadataIdx).toBeGreaterThan(rootKindIdx)
      expect(rootSpecIdx).toBeGreaterThan(rootMetadataIdx)

      // 2. Verify metadata block relative ordering
      const nameIdx = res.indexOf('name: burn-cpu')
      const namespaceIdx = res.indexOf('namespace: chaos-testing')
      const labelsIdx = res.indexOf('labels:')
      const annotationsIdx = res.indexOf('annotations:')

      expect(nameIdx).toBeGreaterThan(rootMetadataIdx)
      expect(namespaceIdx).toBeGreaterThan(nameIdx)
      expect(labelsIdx).toBeGreaterThan(namespaceIdx)
      expect(annotationsIdx).toBeGreaterThan(labelsIdx)

      // 3. Verify nested keys under spec preserve their original order
      const zIdx = res.indexOf('z: z-val')
      const aIdx = res.indexOf('a: a-val')

      expect(zIdx).toBeGreaterThan(rootSpecIdx)
      expect(aIdx).toBeGreaterThan(zIdx) // z was inserted first, so it should stay first

      // 4. Verify embedded resource key ordering (re-rooted sorting)
      const embeddedIdx = res.indexOf('embeddedResource:')
      const embeddedApiVersionIdx = res.indexOf('apiVersion: v1')
      const embeddedKindIdx = res.indexOf('kind: Pod')
      const embeddedMetadataIdx = res.indexOf('name: embedded-name')
      const embeddedSpecIdx = res.indexOf('x: x-val')
      const embeddedSpecBIdx = res.indexOf('b: b-val')

      expect(embeddedIdx).not.toBe(-1)
      expect(embeddedApiVersionIdx).toBeGreaterThan(embeddedIdx)
      expect(embeddedKindIdx).toBeGreaterThan(embeddedApiVersionIdx)
      expect(embeddedMetadataIdx).toBeGreaterThan(embeddedKindIdx)
      expect(embeddedSpecIdx).toBeGreaterThan(embeddedMetadataIdx)
      expect(embeddedSpecBIdx).toBeGreaterThan(embeddedSpecIdx) // x was inserted first, so it should stay first
    })
  })
})
