import { FormattedMessage, IntlShape } from 'react-intl'

// https://github.com/microsoft/TypeScript/issues/24929
function T(id: string): JSX.Element
function T(id: string, intl: IntlShape): string
function T(id: string, intl?: IntlShape) {
  return intl ? intl.formatMessage({ id }) : <FormattedMessage id={id} />
}

export default T
