import CommonTI from 'api/common.type-ti'
import { createCheckers } from 'ts-interface-checker'

const { Config } = createCheckers(CommonTI)

const dummyConfig = {
  security_mode: true,
  dns_server_create: false,
  version: 'xxx',
}

describe('Check common type', () => {
  it('Config', async () => {
    // Normal
    Config.strictCheck(dummyConfig)
  })
})
