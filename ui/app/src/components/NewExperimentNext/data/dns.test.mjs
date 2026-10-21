/*
 * Copyright 2021 Chaos Mesh Authors.
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
import assert from 'node:assert/strict'
import { test } from 'node:test'

import { dnsStaticSchema as staticSchema } from './dns.ts'

test('static DNS accepts IPv4 and IPv6 mappings without patterns', async () => {
  const values = { domainIpMapping: ['Example.COM.=192.0.2.10', 'ipv6.example.com=2001:db8::10'] }
  assert.deepEqual(await staticSchema.validate(values), values)
})

test('static DNS preserves optional pattern filters', async () => {
  const values = { domainIpMapping: ['example.com=192.0.2.10'], patterns: ['example.*'] }
  assert.deepEqual(await staticSchema.validate(values), values)
})

for (const values of [{}, { domainIpMapping: [] }, { domainIpMapping: [''] }]) {
  test(`static DNS rejects missing mappings in ${JSON.stringify(values)}`, async () => {
    await assert.rejects(staticSchema.validate(values))
  })
}

for (const mappings of [
  ['example.com=not-an-ip'],
  ['example.com=999.0.0.1'],
  ['example.com=192.168.01.1'],
  ['example.com=12345::1'],
  ['example.com=fe80::1%eth0'],
  ['example.com=::1]/path'],
  ['*.example.com=192.0.2.1'],
  ['example..com=192.0.2.1'],
  ['example.com..=192.0.2.1'],
  [' example.com=192.0.2.1'],
  [`${'a'.repeat(64)}.com=192.0.2.1`],
  ['example.com'],
  ['example.com=192.0.2.1=bad'],
  ['Example.COM.=192.0.2.1', 'example.com=192.0.2.2'],
]) {
  test(`static DNS rejects invalid mappings ${JSON.stringify(mappings)}`, async () => {
    await assert.rejects(staticSchema.validate({ domainIpMapping: mappings }), { path: 'domainIpMapping' })
  })
}

test('static DNS accepts IPv4-mapped IPv6 and full IPv6 addresses', async () => {
  const values = { domainIpMapping: ['v4.example.com=::ffff:192.0.2.1', 'v6.example.com=2001:0db8:0:0:0:0:0:1'] }
  assert.deepEqual(await staticSchema.validate(values), values)
})
