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

      const lines = res.split('\n')

      // Verify top-level order (only check lines with no leading whitespace)
      const topLevelLines = lines.filter((l) => l.length > 0 && !l.startsWith(' '))
      const topLevelKeys = topLevelLines.map((l) => l.split(':')[0])

      expect(topLevelKeys.indexOf('apiVersion')).toBeLessThan(topLevelKeys.indexOf('kind'))
      expect(topLevelKeys.indexOf('kind')).toBeLessThan(topLevelKeys.indexOf('metadata'))
      expect(topLevelKeys.indexOf('metadata')).toBeLessThan(topLevelKeys.indexOf('spec'))

      // Verify metadata sub-key order by extracting only the metadata block.
      // The metadata block starts after the "metadata:" line and ends at the
      // next top-level key (line with no leading whitespace).
      const metadataLineIdx = lines.findIndex((l) => l === 'metadata:')
      const metadataBlock: string[] = []
      for (let i = metadataLineIdx + 1; i < lines.length; i++) {
        if (lines[i].length > 0 && !lines[i].startsWith(' ')) break
        metadataBlock.push(lines[i].trim())
      }

      const nameIdx = metadataBlock.findIndex((l) => l.startsWith('name:'))
      const namespaceIdx = metadataBlock.findIndex((l) => l.startsWith('namespace:'))
      const labelsIdx = metadataBlock.findIndex((l) => l.startsWith('labels:'))
      const annotationsIdx = metadataBlock.findIndex((l) => l.startsWith('annotations:'))

      expect(nameIdx).not.toBe(-1)
      expect(nameIdx).toBeLessThan(namespaceIdx)
      expect(namespaceIdx).toBeLessThan(labelsIdx)
      expect(labelsIdx).toBeLessThan(annotationsIdx)
    })
  })
})
