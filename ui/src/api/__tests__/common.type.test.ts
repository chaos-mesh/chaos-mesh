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
import CommonTI from 'api/common.type-ti'
import { createCheckers } from 'ts-interface-checker'

const { Config, RBACConfigParams } = createCheckers(CommonTI)

const dummyConfig = {
  security_mode: true,
  dns_server_create: false,
  version: 'xxx',
}

const dummyRBACConfigParams = {
  namespace: 'xxx',
  role: 'viewer',
}

describe('Check common type', () => {
  it('Config', () => {
    // Normal
    Config.strictCheck(dummyConfig)
  })

  it('RBACConfigParams', () => {
    // Normal
    RBACConfigParams.check(dummyRBACConfigParams)

    // Abnormal
    expect(() => RBACConfigParams.check({ ...dummyRBACConfigParams, role: 'xxx' })).toThrow(
      'value.role is none of "manager", "viewer"'
    )
  })
})
