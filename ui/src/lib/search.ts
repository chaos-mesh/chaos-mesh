import { assumeType, difference } from './utils'

import { Archive } from 'api/archives.type'
import { Event } from 'api/events.type'
import { Experiment } from 'api/experiments.type'

type Merge<T extends object, U extends object> = T & U

type MappedEveryFuncToReturnThis<T> = {
  [P in keyof T]: T[P] extends (...args: infer X) => any ? (...args: X) => MappedEveryFuncToReturnThis<T> : T[P]
}

type Keyword = 'namespace' | 'kind' | 'pod' | 'ip' | 'uuid'
export type PropForKeyword = 'experiment' | 'uid' | 'kind' | 'ip' | 'pod' | 'namespace' | 'experiment_id'

export type GlobalSearchData = {
  events: Event[] | []
  experiments: Experiment[] | []
  archives: Archive[] | []
}

export type SearchPath = {
  [k in keyof GlobalSearchData]: { name: PropForKeyword; path: string; value: string }[]
}

interface BaseToken {
  type: 'keyword' | 'content'
  value: string
}

interface KeywordToken extends BaseToken {
  type: 'keyword'
  keyword: Keyword
}

interface ContentToken extends BaseToken {
  type: 'content'
}

type Token = KeywordToken | ContentToken

class ParseSearchAutomata {
  private keywords: Set<Keyword>
  private resolvedKeywords: Set<Keyword>
  private tokens: Token[]

  constructor(keywords: Keyword[]) {
    this.keywords = new Set(keywords)
    this.resolvedKeywords = new Set()
    this.tokens = []
  }

  private get unresolvedKeywords() {
    return difference(this.keywords, this.resolvedKeywords)
  }

  private get parseMethods(): string[] {
    const prototype = Object.getPrototypeOf(this)
    const parseMethods = Object.getOwnPropertyNames(prototype).filter(
      (prop) => prop.startsWith('parse') && prop !== 'parseStart' && prop !== 'parseEnd' && prop !== 'parseMethods'
    )
    // keep parseContent always the last one in parseMethods
    parseMethods.forEach((method, index, array) => {
      if (method === 'parseContent') {
        const tmp = array[index]
        array[index] = array[array.length - 1]
        array[array.length - 1] = tmp
      }
    })
    Object.defineProperty(this, 'parseMethods', {
      value: parseMethods,
      writable: false,
      configurable: true,
    })
    return this.parseMethods
  }

  private emit(token: Token) {
    this.tokens.push(token)
  }

  parseStart(s: string) {
    const prototype = Object.getPrototypeOf(this)
    if (s.length === 0) {
      return this.parseEnd()
    }
    for (let method of this.parseMethods) {
      if (prototype[method].call(this, s) !== false) return this.tokens
    }
    // Since parseContent never return false, so the code below is totally unreachable.
    // It exists just for type check.
    return this.tokens
  }

  private parseEnd() {
    return this.tokens
  }

  private parseKeyword(s: string) {
    for (let keyword of this.unresolvedKeywords) {
      const re = new RegExp(`^(${keyword}):\\s*(\\S+)`)
      const parsedResult = s.match(re)
      if (parsedResult) {
        this.emit({
          type: 'keyword',
          keyword: parsedResult[1],
          value: parsedResult[2],
        } as KeywordToken)
        this.resolvedKeywords.add(parsedResult[1] as Keyword)
        return this.parseStart(s.slice(parsedResult[0].length).trim())
      }
    }
    return false
  }

  private parseContent(s: string) {
    this.emit({
      type: 'content',
      value: s,
    } as ContentToken)
    return this.parseStart('')
  }
}

export function searchGlobal({ events, experiments, archives }: GlobalSearchData, search: string) {
  if (search.length === 0) return {}
  const searchPath: SearchPath = {
    events: [],
    experiments: [],
    archives: [],
  }
  const searchEvents = function (this: GlobalSearchData, keyword: Keyword, value: string) {
    const target = this.events

    if (target.length === 0) return this

    let result: Event[]
    let matchedIndex = -1
    switch (keyword) {
      case 'pod':
        matchedIndex = -1
        result = target.filter((d) => {
          return d.pods?.some((pod, index) => {
            const isMatched = pod.pod_name.toLowerCase().includes(value.toLowerCase())
            if (isMatched) matchedIndex = index
            return isMatched
          })
        })
        searchPath.events.push({
          name: 'pod',
          path: `pods[${matchedIndex}].pod_name`,
          value,
        })
        break
      case 'ip':
        matchedIndex = -1
        result = target.filter((d) => {
          return d.pods?.some((pod, index) => {
            const isMatched = pod.pod_ip.toLowerCase().includes(value.toLowerCase())
            if (isMatched) matchedIndex = index
            return isMatched
          })
        })
        searchPath.events.push({
          name: 'ip',
          path: `pods[${matchedIndex}].pod_ip`,
          value,
        })
        break
      case 'uuid':
        result = target.filter((d) => d.experiment_id.toLowerCase().startsWith(value.toLowerCase()))
        searchPath.events.push({
          name: 'experiment_id',
          path: 'experiment_id',
          value,
        })
        break
      default:
        assumeType<keyof Event & Keyword>(keyword)
        result =
          keyword in target[0]
            ? target.filter((d) => d[keyword]?.toLowerCase().includes(value.toLocaleLowerCase()))
            : []
        searchPath.events.push({
          name: keyword,
          path: keyword,
          value,
        })
        break
    }
    this.events = result
    return this
  }

  const searchExperiments = function (this: GlobalSearchData, keyword: Keyword, value: string) {
    const target = this.experiments

    if (target.length === 0) return this

    let result: Experiment[]
    switch (keyword) {
      case 'uuid':
        result = target.filter((d) => d.uid.toLowerCase().startsWith(value.toLowerCase()))
        searchPath.experiments.push({
          name: 'uid',
          path: 'uid',
          value,
        })
        break
      default:
        assumeType<keyof Experiment & Keyword>(keyword)
        result =
          keyword in target[0] ? target.filter((d) => d[keyword]?.toLowerCase().includes(value.toLowerCase())) : []
        searchPath.experiments.push({
          name: keyword,
          path: keyword,
          value,
        })
        break
    }
    this.experiments = result
    return this
  }

  const searchArchives = function (this: GlobalSearchData, keyword: Keyword, value: string) {
    const target = this.archives

    if (target.length === 0) return this

    let result: Archive[]
    switch (keyword) {
      case 'uuid':
        result = target.filter((d) => d.uid.toLowerCase().startsWith(value.toLowerCase()))
        searchPath.archives.push({
          name: 'uid',
          path: 'uid',
          value,
        })
        break
      default:
        assumeType<keyof Archive & Keyword>(keyword)
        result =
          keyword in target[0] ? target.filter((d) => d[keyword]?.toLowerCase().includes(value.toLowerCase())) : []
        searchPath.archives.push({
          name: keyword,
          path: keyword,
          value,
        })
        break
    }
    this.archives = result
    return this
  }

  const searchForContent = function (this: GlobalSearchData, value: string): GlobalSearchData {
    const { events, experiments, archives } = this
    const eventRes = events.filter((d) => d.experiment.toLowerCase().includes(value.toLowerCase()))
    const experimentRes = experiments.filter((d) => d.name.toLowerCase().includes(value.toLowerCase()))
    const archiveRes = archives.filter((d) => d.name.toLowerCase().includes(value.toLowerCase()))
    searchPath.events.push({
      name: 'experiment',
      path: 'experiment',
      value,
    })
    searchPath.experiments.push({
      name: 'experiment',
      path: 'name',
      value,
    })
    searchPath.archives.push({
      name: 'experiment',
      path: 'name',
      value,
    })
    return {
      events: eventRes,
      experiments: experimentRes,
      archives: archiveRes,
    }
  }

  const protoWithSearchMethods = {
    searchEvents,
    searchExperiments,
    searchArchives,
    searchForContent,
  }

  const keywords: Keyword[] = ['namespace', 'kind', 'pod', 'ip', 'uuid']
  const automata = new ParseSearchAutomata(keywords)
  const tokens = automata.parseStart(search)
  const source = {
    events,
    experiments,
    archives,
  }
  type Result = MappedEveryFuncToReturnThis<Merge<typeof source, typeof protoWithSearchMethods>>
  let result: Result = Object.setPrototypeOf(source, protoWithSearchMethods)

  try {
    tokens.forEach((token) => {
      if (token.type === 'keyword') {
        result = result
          .searchEvents(token.keyword, token.value)
          .searchExperiments(token.keyword, token.value)
          .searchArchives(token.keyword, token.value)
      } else if (token.type === 'content') {
        result = result.searchForContent(token.value)
      }
    })
  } catch (e) {
    ;(result as typeof source) = {
      events: [],
      experiments: [],
      archives: [],
    }
  }
  return {
    searchPath,
    result: result as typeof source,
  }
}
