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
import { CoreWorkflowMeta, TypesArchive, TypesExperiment, TypesSchedule } from '@/openapi/index.schemas'

type Keyword = 'namespace' | 'ns' | 'kind'

interface SearchData {
  workflows: CoreWorkflowMeta[]
  schedules: TypesSchedule[]
  experiments: TypesExperiment[]
  archives: TypesArchive[]
}

interface KeywordToken {
  type: 'keyword'
  keyword: Keyword
  value: string
}

interface ContentToken {
  type: 'content'
  value: string
}

type Token = KeywordToken | ContentToken

class SearchAutomata {
  private keywords: Keyword[]
  private tokens: Token[]

  constructor(keywords: Keyword[]) {
    this.keywords = keywords
    this.tokens = []
  }

  parseStart(s: string) {
    if (s.length === 0) {
      return this.parseEnd()
    }

    return this.parseKeywords(s)
  }

  private parseEnd() {
    return this.tokens
  }

  private parseKeywords(s: string) {
    for (const keyword of this.keywords) {
      const re = new RegExp(`(${keyword}):\\s*(\\S+)`)
      const parsed = s.match(re)

      if (parsed) {
        this.emit({
          type: 'keyword',
          keyword: parsed[1] as Keyword,
          value: parsed[2],
        })

        s = (s.substr(0, parsed.index) + s.substr(parsed.index! + parsed[0].length, s.length)).trim()
      }
    }

    return this.parseContent(s)
  }

  private parseContent(s: string) {
    this.emit({
      type: 'content',
      value: s,
    } as ContentToken)

    return this.parseEnd()
  }

  private emit(token: Token) {
    this.tokens.push(token)
  }

  clearTokens() {
    this.tokens = []
  }
}

const automata = new SearchAutomata(['namespace', 'ns', 'kind'])

export default function search(data: SearchData, s: string) {
  const tokens = automata.parseStart(s)

  const workflows = searchObjects(data.workflows, tokens)
  const schedules = searchObjects(data.schedules, tokens)
  const experiments = searchObjects(data.experiments, tokens)
  const archives = searchObjects(data.archives, tokens)

  automata.clearTokens()

  return { workflows, schedules, experiments, archives }
}

function searchCommon(data: any, keyword: Keyword, value: string) {
  if (keyword === 'ns') {
    keyword = 'namespace'
  }

  return data.filter((d: any) => d[keyword].toLowerCase().includes(value))
}

function searchObjects<T extends { name?: string }>(data: T[], tokens: Token[]) {
  let filtered = data

  tokens.forEach((t) => {
    const val = t.value.toLowerCase()

    if (t.type === 'keyword') {
      filtered = searchCommon(filtered, t.keyword, val)
    } else if (t.type === 'content') {
      filtered = filtered.filter((d) => d.name!.toLowerCase().includes(val))
    }
  })

  return filtered
}
