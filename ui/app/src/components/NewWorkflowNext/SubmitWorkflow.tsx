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
import loadable from '@loadable/component'
import { Box, Divider, MenuItem, Typography } from '@mui/material'
import { Stale } from 'api/queryUtils'
import { Form, Formik } from 'formik'
import yaml from 'js-yaml'
import { useGetCommonChaosAvailableNamespaces, usePostWorkflows } from 'openapi'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import * as Yup from 'yup'

import ConfirmDialog from '@ui/mui-extends/esm/ConfirmDialog'
import Paper from '@ui/mui-extends/esm/Paper'
import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { resetWorkflow } from 'slices/workflows'

import { SelectField, Submit, TextField } from 'components/FormField'
import FormikEffect from 'components/FormikEffect'
import { T } from 'components/T'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

const validationSchema = Yup.object({
  name: Yup.string().trim().required(),
  namespace: Yup.string().trim().required(),
  deadline: Yup.string().trim().required(),
})

interface WorkflowBasic {
  name: string
  namespace: string
  deadline: string
}

interface SubmitWorkflowProps {
  open: boolean
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
  workflow: string
}

export default function SubmitWorkflow({ open, setOpen, workflow }: SubmitWorkflowProps) {
  const navigate = useNavigate()

  const [data, setData] = useState(workflow)
  const [workflowBasic, setWorkflowBasic] = useState<WorkflowBasic>({
    name: '',
    namespace: '',
    deadline: '',
  })

  const { debugMode } = useStoreSelector((state) => state.settings)

  useEffect(() => {
    setData(workflow)
  }, [workflow])

  useEffect(() => {
    setData((oldData) => {
      let { metadata, spec, ...rest }: any = yaml.load(oldData)
      const { name, namespace, deadline } = workflowBasic
      metadata = { ...metadata, name, namespace }

      if (deadline) {
        spec.templates[0].deadline = deadline
      }

      return yaml.dump({ ...rest, metadata, spec })
    })
  }, [workflowBasic])

  const dispatch = useStoreDispatch()

  const { data: namespaces } = useGetCommonChaosAvailableNamespaces({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })
  const { mutateAsync } = usePostWorkflows()

  const submitWorkflow = () => {
    const payload: any = yaml.load(data)

    if (debugMode) {
      console.debug('submitWorkflow => payload', payload)

      return
    }

    mutateAsync({
      data: payload,
    })
      .then(() => {
        dispatch(resetWorkflow())

        navigate('/workflows')
      })
      .catch(console.error)
  }

  return (
    <ConfirmDialog
      open={open}
      close={() => setOpen(false)}
      title="Fill in the basic information and submit"
      dialogProps={{
        PaperProps: {
          style: { width: 1024, height: 768, maxWidth: 'unset' },
        },
      }}
    >
      <Space spacing={6} direction="row" height="100%">
        <Box flexGrow={0} flexShrink={0} flexBasis="45%">
          <Formik
            initialValues={{
              name: '',
              namespace: '',
              deadline: '',
            }}
            validationSchema={validationSchema}
            onSubmit={submitWorkflow}
          >
            {({ errors, touched }) => (
              <>
                <FormikEffect didUpdate={setWorkflowBasic} />
                <Form>
                  <Space>
                    <TextField
                      name="name"
                      label={<T id="common.name" />}
                      helperText={errors.name && touched.name ? errors.name : <T id="newW.nameHelper" />}
                      error={errors.name && touched.name ? true : false}
                    />
                    <SelectField
                      name="namespace"
                      label={<T id="k8s.namespace" />}
                      helperText={
                        errors.namespace && touched.namespace ? errors.namespace : <T id="newE.basic.namespaceHelper" />
                      }
                      error={errors.namespace && touched.namespace ? true : false}
                    >
                      {namespaces!.map((n) => (
                        <MenuItem key={n} value={n}>
                          {n}
                        </MenuItem>
                      ))}
                    </SelectField>
                    <TextField
                      name="deadline"
                      label={<T id="newW.node.deadline" />}
                      helperText={
                        errors.deadline && touched.deadline ? errors.deadline : <T id="newW.node.deadlineHelper" />
                      }
                      error={errors.deadline && touched.deadline ? true : false}
                    />
                    <Submit />
                  </Space>
                </Form>
              </>
            )}
          </Formik>
        </Box>
        <Divider orientation="vertical" flexItem />
        <Space spacing={1.5} flex={1}>
          <Typography variant="body2" fontWeight={500}>
            Preview
          </Typography>
          <Paper sx={{ width: '100%', p: 0 }}>
            <YAMLEditor data={data} />
          </Paper>
        </Space>
      </Space>
    </ConfirmDialog>
  )
}
