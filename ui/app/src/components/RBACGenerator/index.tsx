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
import { Box, Button, Checkbox, FormControl, FormControlLabel, MenuItem, Typography } from '@mui/material'
import { makeStyles } from '@mui/styles'
import api from 'api'
import copy from 'copy-text-to-clipboard'
import { Field, Form, Formik } from 'formik'
import _ from 'lodash'
import { CommonApiCommonRbacConfigGetRequest } from 'openapi'
import { useEffect, useRef, useState } from 'react'
import { useIntl } from 'react-intl'

import Space from '@ui/mui-extends/esm/Space'

import { useStoreDispatch, useStoreSelector } from 'store'

import { setAlert } from 'slices/globalStatus'

import { SelectField } from 'components/FormField'
import i18n from 'components/T'

const useStyles = makeStyles((theme) => ({
  pre: {
    padding: theme.spacing(3),
    background: theme.palette.background.default,
    borderRadius: 4,
    whiteSpace: 'pre-wrap',
  },
  copy: {
    position: 'absolute',
    top: theme.spacing(6),
    right: theme.spacing(3),
  },
}))

const RBACGenerator = () => {
  const classes = useStyles()

  const intl = useIntl()

  const { namespaces } = useStoreSelector((state) => state.experiments)
  const dispatch = useStoreDispatch()

  const [clustered, setClustered] = useState(false)
  const [rbac, setRBAC] = useState('')
  const [getSecret, setGetSecret] = useState('')
  const [generateToken, setGenerateToken] = useState('')
  const containerRef = useRef(null)

  const fetchRBACConfig = (values: CommonApiCommonRbacConfigGetRequest) =>
    api.common.commonRbacConfigGet(values).then(({ data }) => {
      const entries = Object.entries(data)
      const [name, yaml] = entries[0]

      setRBAC(yaml)
      setGetSecret(`kubectl describe${name.includes('cluster') ? '' : ` -n ${values.namespace}`} secrets ${name}`)
      setGenerateToken(`kubectl create token ${name}`)
    })

  useEffect(() => {
    fetchRBACConfig({ namespace: 'default', role: 'viewer' })
  }, [])

  const onValidate = ({ namespace, role, clustered }: CommonApiCommonRbacConfigGetRequest & { clustered: boolean }) => {
    fetchRBACConfig({
      namespace: clustered ? '' : namespace,
      role,
    })
    setClustered(clustered)
  }

  const copyRBAC = () => {
    copy(rbac, { target: containerRef.current! })

    dispatch(
      setAlert({
        type: 'success',
        message: i18n('common.copied', intl),
      })
    )
  }

  return (
    <div ref={containerRef}>
      <Space>
        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.generatorHelper')}
        </Typography>
        <Formik
          initialValues={{ namespace: 'default', role: 'viewer', clustered: false }}
          onSubmit={() => {}}
          validate={onValidate}
          validateOnBlur={false}
        >
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
                {namespaces.map((n) => (
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
        </Formik>
        <Typography variant="body2" color="textSecondary">
          {i18n('settings.addToken.generatorHelper2')}
        </Typography>
        <Box position="relative">
          <pre className={classes.pre} style={{ height: 300, overflow: 'auto' }}>
            {rbac}
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
          <pre className={classes.pre}>{generateToken}</pre>
          <Typography variant="body2" color="textSecondary">
            {i18n('settings.addToken.generatorHelperGetTokenCase2')}
          </Typography>
          <pre className={classes.pre}>{getSecret}</pre>
        </Box>
      </Space>
    </div>
  )
}

export default RBACGenerator
