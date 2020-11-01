import 'jest-expect-message'

import * as TJS from 'typescript-json-schema'
import * as esprima from 'esprima'
import * as mockData from '__mock__/api'

import { assumeType, getAllSubsets, isObject } from 'lib/utils'

import { AxiosResponse } from 'axios'
import Mock from 'mockjs'
import MockAdapter from 'axios-mock-adapter'
import api from 'api'
import { createDebuggerStatement } from 'typescript'
import http from 'api/http'
import { readFileSync } from 'fs'
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

function getGenerator(includedFiles: string[]) {
  const program = TJS.getProgramFromFiles(includedFiles.map((file) => resolve(file)))
  const generator = TJS.buildGenerator(program)
  if (!generator) {
    throw new Error('generator build error')
  }
  return generator
}

function getSchemaByAPIName(includedFiles: string[], apiName: string) {
  const generator = getGenerator(includedFiles)
  const allSymbolNames = generator.getUserSymbols()
  let resSchema: TJS.Definition | undefined
  let paramsSchema: TJS.Definition | undefined
  for (let symbolName of allSymbolNames) {
    const currentSchema = generator.getSchemaForSymbol(symbolName)
    const pattern = /^(\w+):\s*(result|params)\s*/
    const description = currentSchema.description
    const match = description?.match(pattern)
    if (!description || !match) continue
    const [_, currentAPIName, currentMockType] = match
    if (currentAPIName !== apiName) continue
    if (currentMockType === 'result') {
      resSchema = currentSchema
      if (paramsSchema) break
    } else if (currentMockType === 'params') {
      paramsSchema = currentSchema
      if (resSchema) break
    } else {
      throw new Error(`unknown mock type ${currentMockType} for api ${apiName}`)
    }
  }
  return { resSchema, paramsSchema }
}

function getMockByAPIName(apiName: string) {
  const resMock = (mockData as any)[apiName + 'ResMock']
  const paramsMock = (mockData as any)[apiName + 'ParamsMock']
  const requiredParams = (mockData as any)[apiName + 'RequiredParams']

  return {
    resMock,
    paramsMock,
    requiredParams,
  }
}

describe('mock test', () => {
  test('mock data identifier test', () => {
    const mockSuffixes = ['ResMock', 'ParamsMock', 'RequiredParams']

    const filePath = resolve('src/__mock__/api.ts')
    const fileData = readFileSync(filePath, 'utf8')
    const parsedRes = esprima.parseModule(fileData)
    parsedRes.body.forEach((item) => {
      if (item.type === 'ExportNamedDeclaration') {
        if (item.declaration?.type === 'VariableDeclaration') {
          item.declaration.declarations.forEach((declaration) => {
            if (declaration.id.type === 'Identifier') {
              const identifier = declaration.id.name
              const pass = mockSuffixes.some((suffix) => {
                const re = new RegExp(`^\\w*?${suffix}$`)
                return identifier.match(re)
              })
              expect(pass, `The suffix of ${identifier} should be one of ${mockSuffixes}`).toBeTruthy()
            }
          })
        }
      }
    })
  })
})

describe('api test', () => {
  const mock = new MockAdapter(http)

  const allTestedAPI: IndexedTypeByString = {}
  allTestedAPI.archives = api.archives.archives
  allTestedAPI.archiveDetail = api.archives.detail
  allTestedAPI.archiveReport = api.archives.report
  allTestedAPI.namespaces = api.common.namespaces
  allTestedAPI.labels = api.common.labels
  allTestedAPI.annotations = api.common.annotations
  allTestedAPI.pods = api.common.pods
  allTestedAPI.events = api.events.events
  allTestedAPI.dryEvents = api.events.dryEvents
  allTestedAPI.experiments = api.experiments.experiments
  allTestedAPI.newExperiment = api.experiments.newExperiment
  allTestedAPI.startExperiment = api.experiments.startExperiment
  allTestedAPI.deleteExperiment = api.experiments.detail
  allTestedAPI.pauseExperiment = api.experiments.pauseExperiment
  allTestedAPI.updateExperiment = api.experiments.update
  allTestedAPI.experimentDetail = api.experiments.detail

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
    } else if (Object.keys(schema).length === 0) {
      // schema is a empty object, which means the type of the schema is any
      return
    } else if (Array.isArray(schema.type)) {
      // The corresponded type of the schema is a simple union
      const maxFailedNumber = schema.type.length
      let failedNumber = 0
      schema.type.forEach((t) => {
        try {
          testAPIReturnData(apiName, { type: t } as TJS.Definition, data, dataName, definitions)
        } catch (e) {
          failedNumber += 1
        }
        expect(
          failedNumber,
          dataName === apiName
            ? `The return data of the API ${apiName} dose not match any given schema`
            : `The data ${dataName} from the API ${apiName} dose not match any given schema`
        ).toBeLessThan(maxFailedNumber)
      })
    } else if (schema.anyOf) {
      // schema has the anyOf property, which means the corresponded type of the schema is a non-simple union
      const maxFailedNumber = schema.anyOf.length
      let failedNumber = 0
      schema.anyOf.forEach((s) => {
        try {
          testAPIReturnData(apiName, s as TJS.Definition, data, dataName, definitions)
        } catch (e) {
          failedNumber += 1
        }
      })
      expect(
        failedNumber,
        dataName === apiName
          ? `The return data of the API ${apiName} dose not match any given schema`
          : `The data ${dataName} from the API ${apiName} dose not match any given schema`
      ).toBeLessThan(maxFailedNumber)
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
      sendParamsMocks.push([...requiredParamsMock, ...subset])
    })
    return sendParamsMocks
  }

  function startTest(
    apiName: string,
    apiMethod: 'get' | 'post' | 'put' | 'delete',
    apiURL: string,
    includedFiles: string[]
  ) {
    return () => {
      const api: (...args: any) => Promise<AxiosResponse<any>> = allTestedAPI[apiName]

      const mappedMethod = {
        get: 'onGet' as 'onGet',
        post: 'onPost' as 'onPost',
        put: 'onPut' as 'onPut',
        delete: 'onDelete' as 'onDelete',
      }

      const { resSchema, paramsSchema } = getSchemaByAPIName(includedFiles, apiName)
      const { resMock, paramsMock, requiredParams } = getMockByAPIName(apiName)

      mock[mappedMethod[apiMethod]](apiURL, {
        params: {
          asymmetricMatch: (actual: object) => {
            return testAPIParams(apiName, actual, paramsMock, requiredParams)
          },
        },
      }).reply(200, resMock)

      const testAPIReturnDataOrNot = (data: AxiosResponse) => {
        if (resSchema) {
          const properties = resSchema.properties
          if (!properties) {
            console.log('this schema has no property')
            return
          }
          testAPIReturnData(apiName, resSchema, Array.isArray(data.data) ? data.data[0] : data.data)
        }
      }
      if (!paramsSchema) {
        return api().then((data) => {
          testAPIReturnDataOrNot(data)
        })
      } else {
        return Promise.all(
          getSendParamsMocks(paramsSchema).map((mock) =>
            api(...mock).then((data) => {
              testAPIReturnDataOrNot(data)
            })
          )
        )
      }
    }
  }

  describe('archive test', () => {
    const includedFiles = ['src/api/archives.type.ts', 'src/@types/global.d.ts']

    test('archives test', startTest('archives', 'get', 'archives', includedFiles))
    test('archive detail test', startTest('archiveDetail', 'get', 'archives/detail', includedFiles))
    test('archive report test', startTest('archiveReport', 'get', 'archives/report', includedFiles))
  })

  describe('events test', () => {
    const includedFiles = ['src/api/events.type.ts', 'src/@types/global.d.ts']

    test('events test', startTest('events', 'get', 'events', includedFiles))
  })

  describe('experiments test', () => {
    const includedFiles = ['src/api/experiments.type.ts', 'src/@types/global.d.ts']

    test('experiments test', startTest('experiments', 'get', 'experiments', includedFiles))
  })
})
