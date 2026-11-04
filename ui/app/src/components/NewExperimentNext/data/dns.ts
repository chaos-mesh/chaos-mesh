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
import * as Yup from 'yup'

const isIPAddress = (address: string) => {
  if (address.includes(':')) {
    if (!/^[0-9a-fA-F:.]+$/.test(address)) return false
    try {
      return new URL(`http://[${address}]`).hostname.startsWith('[')
    } catch {
      return false
    }
  }

  const octets = address.split('.')
  return octets.length === 4 && octets.every((octet) => /^(0|[1-9][0-9]{0,2})$/.test(octet) && Number(octet) <= 255)
}

export const dnsStaticSchema = Yup.object({
  patterns: Yup.array().of(Yup.string()),
  domainIpMapping: Yup.array()
    .of(Yup.string().required('Each mapping is required'))
    .min(1, 'At least one domain=IP mapping is required')
    .required('At least one domain=IP mapping is required')
    .test('domain-ip-mappings', 'Each mapping must have the form domain=IP', function (entries) {
      const domains = new Set<string>()
      for (const entry of entries ?? []) {
        if (!entry) continue
        const parts = entry.split('=')
        if (parts.length !== 2) return this.createError({ message: 'Each mapping must have the form domain=IP' })
        const [name, address] = parts
        const domain = name.toLowerCase().replace(/\.$/, '')
        if (
          domain.length > 253 ||
          !domain.split('.').every((label) => /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/.test(label))
        ) {
          return this.createError({ message: `Invalid domain name: ${name}` })
        }
        if (!isIPAddress(address)) return this.createError({ message: `Invalid IP address: ${address}` })
        if (domains.has(domain)) return this.createError({ message: `Duplicate domain name: ${name}` })
        domains.add(domain)
      }
      return true
    }),
})
