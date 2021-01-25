import { BasicType, basicTypes } from 'ts-interface-checker/dist/types'

basicTypes['uuid'] = new BasicType((v) => typeof v === 'string', 'is not a UUID')
