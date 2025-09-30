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
import { Box, Button, Checkbox, FormControl, FormControlLabel, MenuItem, Typography } from '@mui/material'
import { styled } from '@mui/material/styles'
import copy from 'copy-text-to-clipboard'
import { Field, Form, Formik } from 'formik'
import _ from 'lodash'
import { useEffect, useRef, useState } from 'react'
import { useIntl } from 'react-intl'

import { SelectField } from '@/components/FormField'
import i18n from '@/components/T'

const PREFIX = 'RBACGenerator'

const classes = {
  pre: `${PREFIX}-pre`,
  copy: `${PREFIX}-copy`,
}

const Root = styled('div')(({ theme }) => ({
  [`& .${classes.pre}`]: {
    padding: theme.spacing(3),
    background: theme.palette.background.default,
    borderRadius: 4,
    whiteSpace: 'pre-wrap',
  },

  [`& .${classes.copy}`]: {
    position: 'absolute',
    top: theme.spacing(6),
    right: theme.spacing(3),
  },
}))

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

      setRBAC({
        yaml,
        getSecret: `kubectl describe${name.includes('cluster') ? '' : ` -n ${params.namespace}`} secrets ${name}`,
        generateToken: `kubectl create token ${name}`,
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

  const copyRBAC = () => {
    if (rbacConfig?.yaml) {
      copy(rbacConfig.yaml, { target: containerRef.current! })

      setAlert({
        type: 'success',
        message: i18n('common.copied', intl),
      })
    }
  }

  return (
    <Root ref={containerRef}>
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
        <Box position="relative">
          <pre className={classes.pre} style={{ height: 300, overflow: 'auto' }}>
            {rbac.yaml}
          </pre>
          <Box className={classes.copy}>
            <Button onClick={copyRBAC}>{i18n('common.copy')}</Button>
          </Box>
        </Box>
        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.generatorHelper3')}
        </Typography>
        <pre className={classes.pre}>kubectl apply -f rbac.yaml</pre>

        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.generatorHelperGetTokenHeader')}
        </Typography>
        <Box position="relative" pl={2}>
          <Typography variant="body2" color="textSecondary">
            {i18n('settings.addToken.generatorHelperGetTokenCase1')}
          </Typography>
          <pre className={classes.pre}>{rbac.generateToken}</pre>
          <Typography variant="body2" color="textSecondary">
            {i18n('settings.addToken.generatorHelperGetTokenCase2')}
          </Typography>
          <pre className={classes.pre}>{rbac.getSecret}</pre>
        </Box>
      </Space>
    </Root>
  )
}

export default RBACGenerator
