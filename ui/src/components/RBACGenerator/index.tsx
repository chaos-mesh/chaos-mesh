import { Box, Checkbox, FormControl, FormControlLabel, MenuItem, Typography } from '@material-ui/core'
import { Field, Form, Formik } from 'formik'
import React, { useEffect, useState } from 'react'

import { RBACConfigParams } from 'api/common.type'
import { SelectField } from 'components/FormField'
import Space from 'components-mui/Space'
import T from 'components/T'
import api from 'api'
import { makeStyles } from '@material-ui/core/styles'
import { toTitleCase } from 'lib/utils'
import { useStoreSelector } from 'store'

const useStyles = makeStyles((theme) => ({
  pre: {
    padding: theme.spacing(3),
    background: theme.palette.background.default,
    borderRadius: 4,
    whiteSpace: 'pre-wrap',
  },
}))

const RBACGenerator = () => {
  const classes = useStyles()

  const { namespaces } = useStoreSelector((state) => state.experiments)

  const [clustered, setClustered] = useState(false)
  const [rbac, setRBAC] = useState('')
  const [getSecret, setGetSecret] = useState('')

  const fetchRBACConfig = (values: RBACConfigParams) =>
    api.common.rbacConfig(values).then(({ data }) => {
      const entries = Object.entries<string>(data)
      const [name, yaml] = entries[0]

      setRBAC(yaml)
      setGetSecret(`kubectl describe${name.includes('cluster') ? '' : ` -n ${values.namespace}`} secrets ${name}`)
    })

  useEffect(() => {
    fetchRBACConfig({ namespace: 'default', role: 'viewer' })
  }, [])

  const onValidate = ({ namespace, role, clustered }: RBACConfigParams & { clustered: boolean }) => {
    fetchRBACConfig({
      namespace: clustered ? '' : namespace,
      role,
    })
    setClustered(clustered)
  }

  return (
    <div>
      <Box mb={3}>
        <Typography variant="body2" color="textSecondary">
          {T('settings.addToken.generatorHelper')}
        </Typography>
      </Box>
      <Formik
        initialValues={{ namespace: 'default', role: 'viewer', clustered: false }}
        onSubmit={() => {}}
        validate={onValidate}
      >
        <Form>
          <Box mb={3}>
            <FormControl>
              <FormControlLabel
                control={<Field as={Checkbox} name="clustered" color="primary" />}
                label={<Typography variant="body2">{T('settings.addToken.clustered')}</Typography>}
              />
            </FormControl>
          </Box>
          <Space display="flex" mb={3}>
            <Box flex={1}>
              <SelectField
                name="namespace"
                label={T('newE.basic.namespace')}
                helperText={T('common.chooseNamespace')}
                disabled={clustered}
              >
                {namespaces.map((n) => (
                  <MenuItem key={n} value={n}>
                    {n}
                  </MenuItem>
                ))}
              </SelectField>
            </Box>
            <Box flex={1}>
              <SelectField
                name="role"
                label={T('settings.addToken.role')}
                helperText={T('settings.addToken.roleHelper')}
              >
                {['manager', 'viewer'].map((role) => (
                  <MenuItem key={role} value={role}>
                    {toTitleCase(role)}
                  </MenuItem>
                ))}
              </SelectField>
            </Box>
          </Space>
        </Form>
      </Formik>
      <Typography variant="body2" color="textSecondary">
        {T('settings.addToken.generatorHelper2')}
      </Typography>
      <pre className={classes.pre} style={{ height: 750 }}>
        {rbac}
      </pre>
      <Typography variant="body2" color="textSecondary">
        {T('settings.addToken.generatorHelper3')}
      </Typography>
      <pre className={classes.pre}>kubectl apply -f rbac.yaml</pre>
      <Typography variant="body2" color="textSecondary">
        {T('settings.addToken.generatorHelper4')}
      </Typography>
      <pre className={classes.pre}>{getSecret}</pre>
    </div>
  )
}

export default RBACGenerator
