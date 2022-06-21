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
import { Form, Formik } from 'formik'
import { Submit, TextField } from 'components/FormField'
import { Template, TemplateType } from 'slices/workflows'
import { validateDeadline, validateName } from 'lib/formikhelpers'

import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import Space from '@ui/mui-extends/esm/Space'
import i18n from 'components/T'
import { schemaBasic } from './types'
import { useIntl } from 'react-intl'

export interface SuspendValues {
  name: string
  deadline: string
}

interface SuspendProps {
  initialValues?: SuspendValues
  submit: (template: Template) => void
}

const Suspend: React.FC<SuspendProps> = ({ initialValues, submit }) => {
  const intl = useIntl()

  const onSubmit = ({ name, deadline }: SuspendValues) => {
    const values = schemaBasic.cast({ name, deadline })

    submit({
      type: TemplateType.Suspend,
      ...values!,
    })
  }

  return (
    <Paper>
      <Space>
        <PaperTop title={i18n('newW.suspendTitle')} />
        <Formik initialValues={initialValues || { name: '', deadline: '' }} onSubmit={onSubmit}>
          {({ errors, touched }) => (
            <Form>
              <Space>
                <TextField
                  fast
                  name="name"
                  label={i18n('common.name')}
                  validate={validateName(i18n('newW.node.nameValidation', intl))}
                  helperText={errors.name && touched.name ? errors.name : i18n('newW.node.nameHelper')}
                  error={errors.name && touched.name ? true : false}
                />
                <TextField
                  fast
                  name="deadline"
                  label={i18n('newW.node.deadline')}
                  validate={validateDeadline(i18n('newW.node.deadlineValidation', intl))}
                  helperText={errors.deadline && touched.deadline ? errors.deadline : i18n('newW.node.deadlineHelper')}
                  error={errors.deadline && touched.deadline ? true : false}
                />
              </Space>
              <Submit />
            </Form>
          )}
        </Formik>
      </Space>
    </Paper>
  )
}

export default Suspend
