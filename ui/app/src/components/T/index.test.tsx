/*
 * Copyright 2022 Chaos Mesh Authors.
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
import { render, screen } from 'test-utils'

import i18n, { T } from '.'

describe('i18n() and <T />', () => {
  test('displays `k8s.title` with i18n()', async () => {
    render(<div>{i18n('k8s.title')}</div>)

    expect(screen.getByText('Kubernetes')).toBeInTheDocument()
  })

  test('displays `k8s.title` with <T />', async () => {
    render(<T id="k8s.title" />)

    expect(screen.getByText('Kubernetes')).toBeInTheDocument()
  })
})
