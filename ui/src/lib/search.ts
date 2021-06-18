import { Archive } from 'api/archives.type'
import { Experiment } from 'api/experiments.type'

type Keyword = 'namespace' | 'kind'

interface SearchData {
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

const automata = new SearchAutomata(['namespace', 'kind'])

export default function search(data: SearchData, s: string) {
  const tokens = automata.parseStart(s)

  const experiments = searchExperiments(data.experiments, tokens)
  const archives = searchExperiments<Archive>(data.archives, tokens)

  automata.clearTokens()

  return { experiments, archives }
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
