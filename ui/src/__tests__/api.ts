import * as TJS from 'typescript-json-schema'

import {
  archivesParamsMock,
  archivesRequiredParams,
  archivesResMock,
  eventsParamsMock,
  eventsRequiredParams,
  eventsResMock,
} from '__mock__/api'

import MockAdapter from 'axios-mock-adapter'
import api from 'api'
import http from 'api/http'
import { resolve } from 'path'
import { send } from 'process'

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

  describe('archive test', () => {
    const program = TJS.getProgramFromFiles([resolve('src/api/archives.type.ts'), resolve('src/@types/global.d.ts')])
    const generator = TJS.buildGenerator(program)
    if (!generator) {
      throw new Error('generator build error')
    }
    const schema = generator.getSchemaForSymbol('Archive')
    const foo = generator.getSchemaForSymbol('GetArchivesParams')
    console.log(foo)

    mock
      .onGet('archives', {
        params: {
          asymmetricMatch: (actual: any) => {
            console.log(actual)
            const sendParams = Object.keys(actual)
            const apiParams = Object.keys(archivesParamsMock)
            expect(send).toEqual(expect.arrayContaining(archivesRequiredParams))
            expect(apiParams).toEqual(expect.arrayContaining(sendParams))
            return true
          },
        },
      })
      .reply(200, archivesResMock)

    const { namespace, name, kind } = archivesParamsMock
    test('archive test', () => {
      return archives(namespace, name, kind).then((data) => {
        const properties = schema.properties

        if (!properties) {
          console.log('this schema has no property')
          return
        }

        Object.keys(properties).forEach((prop) => {
          const type = (properties[prop] as TJS.Definition).type
          expect(data.data[0]).toEqual(
            expect.objectContaining({
              [prop]: expect.anything(),
            })
          )
          expect(typeof (data.data[0] as any)[prop]).toEqual(type)
        })
      })
    })
  })

  describe('events test', () => {
    const program = TJS.getProgramFromFiles([resolve('src/api/events.type.ts'), resolve('src/@types/global.d.ts')])
    const generator = TJS.buildGenerator(program)
    if (!generator) {
      throw new Error('generator build error')
    }
    const schema = generator.getSchemaForSymbol('Event')
    console.log(schema)

    mock
      .onGet('events', {
        params: {
          asymmetricMatch: (actual: any) => {
            const sendParams = Object.keys(actual)
            const apiParams = Object.keys(eventsParamsMock)
            expect(send).toEqual(expect.arrayContaining(eventsRequiredParams))
            expect(apiParams).toEqual(expect.arrayContaining(sendParams))
            return true
          },
        },
      })
      .reply(200, eventsResMock)

    test('events test', () => {
      return events(eventsParamsMock).then((data) => {
        const properties = schema.properties

        if (!properties) {
          console.log('this schema has no property')
          return
        }

        Object.keys(properties).forEach((prop) => {
          const type = (properties[prop] as TJS.Definition).type
          expect(data.data[0]).toEqual(
            expect.objectContaining({
              [prop]: expect.anything(),
            })
          )
          expect(typeof (data.data[0] as any)[prop]).toEqual(type)
        })
      })
    })
  })
})
