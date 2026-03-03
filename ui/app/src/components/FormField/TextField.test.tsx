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
import { Formik } from 'formik'
import { act, fireEvent, render, screen } from 'test-utils'

import TextField from './TextField'

const handleSubmit = jest.fn()

describe('TextField', () => {
  it('should prevent wheel on number field', async () => {
    render(
      <Formik initialValues={{}} onSubmit={handleSubmit}>
        <TextField type="number" name="myfield" inputProps={{ 'data-testid': 'input' }} />
      </Formik>
    )
    screen.queryByTestId('input')?.focus()

    expect(screen.queryByTestId('input')).toBeTruthy()

    await act(async () => {
      fireEvent.wheel(screen.getByTestId('input'))
    })

    expect(screen.getByTestId('input')).not.toHaveFocus()
  })

  it('should prevent wheel on non-number field', async () => {
    render(
      <Formik initialValues={{}} onSubmit={handleSubmit}>
        <TextField name="myfield" inputProps={{ 'data-testid': 'input' }} />
      </Formik>
    )
    screen.queryByTestId('input')?.focus()

    expect(screen.queryByTestId('input')).toBeTruthy()

    await act(async () => {
      fireEvent.wheel(screen.getByTestId('input'))
    })

    expect(screen.getByTestId('input')).toHaveFocus()
  })
})
