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
