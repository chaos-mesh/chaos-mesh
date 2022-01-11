import { genForms } from './index.js'
import { hideBin } from 'yargs/helpers'
import yargs from 'yargs'

const argv = yargs(hideBin(process.argv))
  .command('formik', 'convert CRDs to TypeScript forms with @openapitools/openapi-generator-cli generated')
  .alias('help', 'h')
  .version(false)
  .wrap(120).argv

// eslint-disable-next-line default-case
switch (argv._[0]) {
  case 'formik':
    runFormik()

    break
}

/**
 * Internal function to convert CRDs to TypeScript forms.
 *
 */
function runFormik() {
  genForms('../../app/src/openapi/api.ts')
}
