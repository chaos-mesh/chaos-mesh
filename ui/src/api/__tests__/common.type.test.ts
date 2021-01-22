import CommonTI from 'api/common.type-ti'
import { createCheckers } from 'ts-interface-checker'

const { Config } = createCheckers(CommonTI)

const dummyConfig = {
  security_mode: true,
}

describe('Check common type', () => {
  it('Config', async () => {
    // Normal
    Config.strictCheck(dummyConfig)
  })
})
