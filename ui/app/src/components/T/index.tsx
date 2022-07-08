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

/*
 * This component was created to provide localized translations.
 *
 * It has three ways to use:
 *
 * 1. **DEPRECATED** `i18n(string)` will return a `FormattedMessage` component with the given string as id.
 * 2. `i18n(string, intl)` will use intl object to return a translated string.
 * 3. `T` is an alias of `FormattedMessage`. Mostly you will often use `T` instead of `i18n`.
 *
 */
import { FormattedMessage } from 'react-intl'
import type { IntlShape } from 'react-intl'

// https://github.com/microsoft/TypeScript/issues/24929
function i18n(id: string): Exclude<React.ReactChild, number> // DEPRECATED, but preserve for backward compatibility.
function i18n(id: string, intl: IntlShape): string
function i18n(id: string, intl?: IntlShape) {
  return intl ? intl.formatMessage({ id }) : <FormattedMessage id={id} />
}

// Re-export an alias for FormattedMessage.
export const T = FormattedMessage
export default i18n
