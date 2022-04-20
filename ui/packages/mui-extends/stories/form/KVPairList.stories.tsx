import { ComponentMeta, ComponentStory } from '@storybook/react'
import { Field, Form, Formik, FormikProps } from 'formik'
import React, { useState } from 'react'

import KVPairList from '../../esm/KVPairList'
import type { KVPairListProps } from '../../esm/KVPairList'

export default {
  title: 'Form/KVPairList',
  component: KVPairList,
  argTypes: {
    onChange: {
      action: 'onChange',
    },
  },
} as ComponentMeta<typeof KVPairList>

const Template: ComponentStory<typeof KVPairList> = ({ ...props }: KVPairListProps) => {
  return <KVPairList {...props} />
}

export const Default = Template.bind({})
Default.args = {
  name: 'spec.abort',
  label: 'HTTP Headers',
  helperText: 'Set HTTP Headers',
  helperTextForKey: '',
  helperTextForValue: '',
  disabled: false,
  error: false,
  initialData: [
    { key: 'key-1', value: 'value-1' },
    { key: 'key-2', value: 'value-2' },
    { key: 'key-3', value: 'value-3' },
  ],
}

const FormikTemplate: ComponentStory<typeof KVPairList> = ({ ...props }: any) => {
  return (
    <div>
      <Formik initialValues={props.formData} onSubmit={(values) => {}}>
        {(formikProps: FormikProps<any>) => {
          return (
            <div>
              <Form>
                <Field name={props.name}>
                  {({ field, form, meta }) => {
                    console.log(field)
                    console.log(form)
                    console.log(meta)
                    const value = formikProps.getFieldProps(props.name).value
                    console.log('render value')
                    console.log(value)
                    return (
                      <div>
                        <KVPairList
                          label="HTTPHeader"
                          name={field.name}
                          value={value}
                          onChange={field.onChange}
                        ></KVPairList>
                      </div>
                    )
                  }}
                </Field>
              </Form>
            </div>
          )
        }}
      </Formik>
    </div>
  )
}

export const WithFormik = FormikTemplate.bind({})
WithFormik.args = {
  name: 'spec.abort',
  label: 'HTTP Headers',
  helperText: 'Set HTTP Headers',
  helperTextForKey: '',
  helperTextForValue: '',
  disabled: false,
  error: false,
  formData: {
    spec: {
      abort: {
        k1: 'v1',
        k2: 'v2',
      },
    },
  },
}
