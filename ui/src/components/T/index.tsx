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
import { FormattedMessage, IntlShape } from 'react-intl'

// https://github.com/microsoft/TypeScript/issues/24929
function T(id: string): JSX.Element
function T(id: string, intl: IntlShape): string
function T(id: string, intl?: IntlShape) {
  return intl ? intl.formatMessage({ id }) : <FormattedMessage id={id} />
}

export default T
