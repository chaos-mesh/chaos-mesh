import esprima from 'esprima'
import { readFileSync } from 'fs'
import { resolve } from 'path'

const filePath = ''
const fileData = readFileSync(resolve('src/__mock/__/api.ts'), 'utf8')

console.log(esprima.parseModule(fileData))
