import { Archive } from 'api/archives.type'
import { Event } from 'api/events.type'
import { Experiment } from 'api/experiments.type'

type Keyword = 'namespace' | 'kind' | 'pod' | 'ip'

interface SearchData {
  events: Event[]
  experiments: Experiment[]
  archives: Archive[]
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

const automata = new SearchAutomata(['namespace', 'kind', 'pod', 'ip'])

export default function search(data: SearchData, s: string) {
  const tokens = automata.parseStart(s)

  const experiments = searchExperiments(data.experiments, tokens)
  const events = searchEvents(data.events, tokens)
  const archives = searchExperiments<Archive>(data.archives, tokens)

  automata.clearTokens()

  return { experiments, events, archives }
}

function searchCommon(data: any, keyword: string, value: string) {
  if (keyword === 'namespace' || keyword === 'kind') {
    return data.filter((d: any) => d[keyword].toLowerCase().includes(value))
  }

  return data
}

function searchExperiments<T extends { name: string } = Experiment>(data: T[], tokens: Token[]) {
  let filtered = data

  tokens.forEach((t) => {
    const val = t.value.toLowerCase()

    if (t.type === 'keyword') {
      filtered = searchCommon(filtered, t.keyword, val)
    } else if (t.type === 'content') {
      filtered = filtered.filter((d) => d.name.toLowerCase().includes(val))
    }
  })

  return filtered
}

function searchEvents(data: SearchData['events'], tokens: Token[]) {
  let filtered = data

  tokens.forEach((t) => {
    const val = t.value.toLowerCase()

    if (t.type === 'keyword') {
      switch (t.keyword) {
        case 'pod':
          filtered = filtered.filter((d) => d.pods?.some((pod) => pod.pod_name.toLowerCase().includes(val)))
          break
        case 'ip':
          filtered = filtered.filter((d) => d.pods?.some((pod) => pod.pod_ip.includes(val)))
          break
        default:
          filtered = searchCommon(filtered, t.keyword, val)
          break
      }
    } else if (t.type === 'content') {
      filtered = filtered.filter((d) => d.experiment.toLowerCase().includes(val))
    }
  })

  return filtered
}
