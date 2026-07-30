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
import { Stale } from '@/api/queryUtils'
import Space from '@/mui-extends/Space'
import { useGetCommonChaosAvailableNamespaces, useGetCommonRbacConfig } from '@/openapi'
import { useComponentActions } from '@/zustand/component'
import { Box, Checkbox, FormControl, FormControlLabel, MenuItem, Typography } from '@mui/material'
import copy from 'copy-text-to-clipboard'
import { Field, Form, Formik } from 'formik'
import _ from 'lodash'
import { useEffect, useRef, useState } from 'react'
import { useIntl } from 'react-intl'

import { SelectField } from '@/components/FormField'
import i18n from '@/components/T'

import CopyableCodeBlock from './CopyableCodeBlock'

const initialValues = { namespace: 'default', role: 'viewer', clustered: false }

const RBACGenerator = () => {
  const intl = useIntl()

  const { setAlert } = useComponentActions()

  const [params, setParams] = useState(initialValues)
  const [rbac, setRBAC] = useState({
    yaml: '',
    getSecret: '',
    generateToken: '',
  })
  const containerRef = useRef(null)

  const { data: namespaces } = useGetCommonChaosAvailableNamespaces({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })
  const { data: rbacConfig } = useGetCommonRbacConfig(params)

  useEffect(() => {
    if (rbacConfig) {
      const [name, yaml] = Object.entries(rbacConfig)[0]

      const isClusterScoped = name.includes('cluster')
      const serviceAccountNamespace = isClusterScoped ? 'default' : params.namespace

      const describeNamespaceArg = !isClusterScoped && params.namespace ? ` -n ${params.namespace}` : ''
      const tokenNamespaceArg = serviceAccountNamespace ? ` -n ${serviceAccountNamespace}` : ''

      setRBAC({
        yaml,
        getSecret: `kubectl describe${describeNamespaceArg} secrets ${name}`,
        generateToken: `kubectl${tokenNamespaceArg} create token ${name}`,
      })
    }
  }, [rbacConfig, params])

  const onValidate = ({ namespace, role, clustered }: typeof params) => {
    setParams({
      namespace: clustered ? '' : namespace,
      role,
      clustered,
    })
  }

  const copyCommand = (text: string) => {
    if (text && copy(text, { target: containerRef.current! })) {
      setAlert({
        type: 'success',
        message: i18n('common.copied', intl),
      })
    }
  }

  return (
    <div ref={containerRef}>
      <Space>
        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.generatorHelper')}
        </Typography>
        <Formik initialValues={initialValues} onSubmit={() => {}} validate={onValidate} validateOnBlur={false}>
          {({ values: { clustered } }) => (
            <Form>
              <Space>
                <FormControl>
                  <FormControlLabel
                    control={<Field as={Checkbox} name="clustered" color="primary" />}
                    label={<Typography variant="body2">{i18n('settings.addToken.clustered')}</Typography>}
                  />
                </FormControl>
                <SelectField
                  name="namespace"
                  label={i18n('k8s.namespace')}
                  helperText={i18n('common.chooseNamespace')}
                  disabled={clustered}
                >
                  {namespaces!.map((n) => (
                    <MenuItem key={n} value={n}>
                      {n}
                    </MenuItem>
                  ))}
                </SelectField>
                <SelectField
                  name="role"
                  label={i18n('settings.addToken.role')}
                  helperText={i18n('settings.addToken.roleHelper')}
                >
                  {['manager', 'viewer'].map((role) => (
                    <MenuItem key={role} value={role}>
                      {_.upperFirst(role)}
                    </MenuItem>
                  ))}
                </SelectField>
              </Space>
            </Form>
          )}
        </Formik>
        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.generatorHelper2')}
        </Typography>
        <CopyableCodeBlock text={rbac.yaml} copyLabel={i18n('common.copy')} onCopy={copyCommand} height={300} />
        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.generatorHelper3')}
        </Typography>

        <CopyableCodeBlock
          text="kubectl apply -f rbac.yaml"
          copyLabel={i18n('common.copy')}
          onCopy={copyCommand}
          singleLine
        />

        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.generatorHelperGetTokenHeader')}
        </Typography>
        <Box position="relative" pl={2}>
          <Typography variant="body2" color="textSecondary">
            {i18n('settings.addToken.generatorHelperGetTokenCase1')}
          </Typography>
          <CopyableCodeBlock
            text={rbac.generateToken}
            copyLabel={i18n('common.copy')}
            onCopy={copyCommand}
            singleLine
          />

          <Typography variant="body2" color="textSecondary">
            {i18n('settings.addToken.generatorHelperGetTokenCase2')}
          </Typography>
          <CopyableCodeBlock text={rbac.getSecret} copyLabel={i18n('common.copy')} onCopy={copyCommand} singleLine />
        </Box>
      </Space>
    </div>
  )
}

export default RBACGenerator
