import 'jest-expect-message'

import * as TJS from 'typescript-json-schema'

import {
  archivesParamsMock,
  archivesRequiredParams,
  archivesResMock,
  eventsParamsMock,
  eventsRequiredParams,
  eventsResMock,
} from '__mock__/api'
import { assumeType, getAllSubsets, isObject } from 'lib/utils'

import { AxiosResponse } from 'axios'
import Mock from 'mockjs'
import MockAdapter from 'axios-mock-adapter'
import api from 'api'
import http from 'api/http'
import { resolve } from 'path'

const Random = Mock.Random

function getSchemaPropType(x: unknown) {
  if (isObject(x)) {
    return 'object'
  } else if (Array.isArray(x)) {
    return 'array'
  } else {
    return typeof x
  }
}

function getDefinitionByRef(ref: string) {
  return ref.split('/').slice(-1)[0]
}

function getMockBySchema(
  schema: TJS.Definition,
  definitions?: {
    [key: string]: TJS.DefinitionOrBoolean
  }
): any {
  if (schema.type) {
    switch (schema.type) {
      case 'string':
        return Random.string()
      case 'number':
        return Random.natural()
      case 'boolean':
        return Random.boolean()
      case 'object':
        const mockSource = Object.keys(schema.properties!).reduce((mockObj, prop) => {
          mockObj[prop] = getMockBySchema(schema.properties![prop] as TJS.Definition, definitions)
          return mockObj
        }, {} as IndexedTypeByString)
        return Mock.mock(mockSource)
      case 'array':
        /* Since the array type has a property named minItem, which means how many items are required in the array.
           So there are multiple cases to mock an array, and the logic is over complex and unnecessary for this function.
        */
        break
      default:
        throw new Error(`unknown type: ${schema.type}`)
    }
  } else if (schema.$ref) {
    const def = getDefinitionByRef(schema.$ref)
    return getMockBySchema(definitions![def] as TJS.Definition, definitions)
  }
}

function buildSchema(includedFiles: string[], paramsInterfaceName: string, resInterfaceName: string) {
  const program = TJS.getProgramFromFiles(includedFiles.map((file) => resolve(file)))
  const generator = TJS.buildGenerator(program)
  if (!generator) {
    throw new Error('generator build error')
  }
  const resSchema = generator.getSchemaForSymbol(resInterfaceName)
  const paramsSchema = generator.getSchemaForSymbol(paramsInterfaceName)
  return { resSchema, paramsSchema }
}

describe('api test', () => {
  const mock = new MockAdapter(http)
  const archives = api.archives.archives
  const archiveDetail = api.archives.detail
  const report = api.archives.report
  const namespaces = api.common.namespaces
  const labels = api.common.labels
  const annotations = api.common.annotations
  const pods = api.common.pods
  const events = api.events.events
  const dryEvents = api.events.dryEvents
  const experiments = api.experiments.experiments
  const newExperiment = api.experiments.newExperiment
  const startExperiment = api.experiments.startExperiment
  const deleteExperiment = api.experiments.detail
  const pauseExperiment = api.experiments.pauseExperiment
  const updateExperiment = api.experiments.update
  const experimentDetail = api.experiments.detail

  function testAPIReturnData(
    apiName: string,
    schema: TJS.Definition,
    data: any,
    dataName: string = apiName,
    definitions: IndexedTypeByString = {}
  ) {
    if (schema.type === 'object' && schema.properties) {
      if (schema.definitions) {
        Object.assign(definitions, schema.definitions)
      }
      Object.entries(schema.properties).forEach(([prop, val]) => {
        expect(
          data,
          dataName === apiName
            ? `The return data of the API ${apiName} should contain ${prop}`
            : `The data ${dataName} from the API ${apiName} should contain ${prop}`
        ).toEqual(
          expect.objectContaining({
            [prop]: expect.anything(),
          })
        )
        testAPIReturnData(apiName, val as TJS.Definition, data[prop], prop, definitions)
      })
    } else if (schema.type === 'array' && schema.items) {
      expect(
        Array.isArray(data),
        dataName === apiName
          ? `The return data of the API ${apiName} should be an array`
          : `The data ${dataName} from the API ${apiName} should be an array`
      ).toBeTruthy()
      testAPIReturnData(apiName, schema.items as TJS.Definition, data[0], dataName + "'s item", definitions)
    } else if (schema.$ref) {
      const def = getDefinitionByRef(schema.$ref)
      testAPIReturnData(apiName, definitions[def], data, dataName, definitions)
    } else {
      expect(
        schema.type,
        dataName === apiName
          ? `The type of data the API ${apiName} returns should be ${schema.type}`
          : `The type of data ${dataName} from the API ${apiName} should be ${schema.type}`
      ).toEqual(getSchemaPropType(data))
    }
  }

  function testAPIParams(
    apiName: string,
    sendParams: IndexedTypeByString | undefined,
    apiParams: IndexedTypeByString | undefined,
    apiRequiredParams: string[]
  ) {
    const sendParamsList = (sendParams && Object.keys(sendParams)) || []
    const apiParamsList = (apiParams && Object.keys(apiParams)) || []
    expect(
      sendParamsList,
      `The API ${apiName} parameters are ${
        sendParamsList.length === 0 ? null : sendParamsList
      } now, but still require ${apiRequiredParams}`
    ).toEqual(expect.arrayContaining(apiRequiredParams))
    expect(
      apiParamsList,
      `The API ${apiName} parameters are ${sendParamsList} now, but we just need parameters included by ${
        apiParamsList.length === 0 ? null : apiParamsList
      }`
    ).toEqual(expect.arrayContaining(sendParamsList))
    sendParamsList.forEach((param) => {
      const sendParamType = typeof sendParams![param]
      const apiParamType = typeof apiParams![param]
      expect(
        sendParamType,
        `The type of ${param} from the API ${apiName} is ${sendParamType}, but we need ${apiParamType}`
      ).toEqual(apiParamType)
    })
    return true
  }

  function getSendParamsMocks(schema: TJS.Definition) {
    if (schema.type !== 'array' || !schema.items) {
      throw new Error('wrong schema format')
    }
    assumeType<TJS.Definition[]>(schema.items)
    const requiredParams = schema.items.slice(0, schema.minItems)
    const optionalParams = schema.items.slice(schema.minItems)
    const requiredParamsMock: any[] = []
    const sendParamsMocks: any[][] = []

    for (let i = 0; i < schema.minItems!; i++) {
      requiredParamsMock.push(getMockBySchema(requiredParams[i], schema.definitions))
    }

    getAllSubsets(optionalParams.map((param) => getMockBySchema(param, schema.definitions))).forEach((subset) => {
      sendParamsMocks.push([...requiredParams, ...subset])
    })

    return sendParamsMocks
  }

  function startTest(
    apiName: string,
    api: (...args: any) => Promise<AxiosResponse<any>>,
    paramsSchema: TJS.Definition,
    resSchema: TJS.Definition
  ) {
    return () => {
      return Promise.all(
        getSendParamsMocks(paramsSchema).map((mock) =>
          api(...mock).then((data) => {
            const properties = resSchema.properties
            if (!properties) {
              console.log('this schema has no property')
              return
            }
            testAPIReturnData(apiName, resSchema, data.data[0])
          })
        )
      )
    }
  }

  describe('archive test', () => {
    const { paramsSchema, resSchema } = buildSchema(
      ['src/api/archives.type.ts', 'src/@types/global.d.ts'],
      'GetArchivesParams',
      'Archive'
    )
    mock
      .onGet('archives', {
        params: {
          asymmetricMatch: (actual: object) => {
            return testAPIParams('archives', actual, archivesParamsMock, archivesRequiredParams)
          },
        },
      })
      .reply(200, archivesResMock)

    test('archives test', startTest('archives', archives, paramsSchema, resSchema))

    describe('events test', () => {
      const { paramsSchema, resSchema } = buildSchema(
        ['src/api/events.type.ts', 'src/@types/global.d.ts'],
        'GetEventsParams',
        'Event'
      )

      mock
        .onGet('events', {
          params: {
            asymmetricMatch: (actual: object) => {
              return testAPIParams('events', actual, eventsParamsMock, eventsRequiredParams)
            },
          },
        })
        .reply(200, eventsResMock)

      test('events test', startTest('events', events, paramsSchema, resSchema))
    })
  })
})
