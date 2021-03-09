import { Box, Button, Divider, Grid, MenuItem, Typography } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, TextField } from 'components/FormField'
import React, { useEffect, useState } from 'react'
import basicData, { schema } from './data/basic'
import { createStyles, makeStyles } from '@material-ui/core/styles'
import { setBasic, setStep2 } from 'slices/experiments'
import { useStoreDispatch, useStoreSelector } from 'store'

import AdvancedOptions from 'components/AdvancedOptions'
import CheckCircleOutlineIcon from '@material-ui/icons/CheckCircleOutline'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import PublishIcon from '@material-ui/icons/Publish'
import Scheduler from './form/Scheduler'
import Scope from './form/Scope'
import SkeletonN from 'components-mui/SkeletonN'
import T from 'components/T'
import UndoIcon from '@material-ui/icons/Undo'

const useStyles = makeStyles((theme) =>
  createStyles({
    submit: {
      borderColor: theme.palette.success.main,
    },
    submitIcon: {
      color: theme.palette.success.main,
    },
    asButton: {
      cursor: 'pointer',
    },
  })
)

const Step2 = () => {
  const classes = useStyles()

  const { namespaces, step2, basic } = useStoreSelector((state) => state.experiments)
  const dispatch = useStoreDispatch()

  const [init, setInit] = useState(basicData)

  useEffect(() => {
    setInit({
      ...basicData,
      ...basic,
    })
  }, [basic])

  const handleOnSubmitStep2 = (values: Record<string, any>) => {
    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug handleSubmitStep2', values)
    }

    dispatch(setBasic(values))
    dispatch(setStep2(true))
  }

  const handleUndo = () => dispatch(setStep2(false))

  return (
    <Paper className={step2 ? classes.submit : ''}>
      <PaperTop
        title={
          <Box display="flex">
            {step2 && (
              <Box display="flex" alignItems="center" mr={3}>
                <CheckCircleOutlineIcon className={classes.submitIcon} />
              </Box>
            )}
            {T('newE.titleStep2')}
          </Box>
        }
      >
        {step2 && (
          <Box display="flex" alignItems="center">
            <UndoIcon className={classes.asButton} onClick={handleUndo} />
          </Box>
        )}
      </PaperTop>
      <Box position="relative" p={6} hidden={step2}>
        <Formik
          enableReinitialize
          initialValues={init}
          validationSchema={schema}
          validateOnChange={false}
          onSubmit={handleOnSubmitStep2}
        >
          {({ errors, touched }) => (
            <Form>
              <Grid container spacing={9}>
                <Grid item xs={12} md={6}>
                  <Box mb={3}>
                    <Typography>{T('newE.steps.scope')}</Typography>
                  </Box>
                  {namespaces.length ? <Scope namespaces={namespaces} /> : <SkeletonN n={6} />}
                </Grid>
                <Grid item xs={12} md={6}>
                  <Box mb={3}>
                    <Typography>{T('newE.steps.basic')}</Typography>
                  </Box>
                  <TextField
                    fast
                    name="name"
                    label={T('newE.basic.name')}
                    helperText={errors.name && touched.name ? errors.name : T('newE.basic.nameHelper')}
                    error={errors.name && touched.name ? true : false}
                  />

                  <AdvancedOptions>
                    {namespaces.length && (
                      <SelectField
                        name="namespace"
                        label={T('newE.basic.namespace')}
                        helperText={T('newE.basic.namespaceHelper')}
                      >
                        {namespaces.map((n) => (
                          <MenuItem key={n} value={n}>
                            {n}
                          </MenuItem>
                        ))}
                      </SelectField>
                    )}
                    <LabelField name="labels" label={T('k8s.labels')} isKV />
                    <LabelField name="annotations" label={T('k8s.annotations')} isKV />
                  </AdvancedOptions>
                  <Box mb={3}>
                    <Divider />
                  </Box>
                  <Scheduler errors={errors} touched={touched} />
                  <Box mt={6} textAlign="right">
                    <Button type="submit" variant="contained" color="primary" startIcon={<PublishIcon />}>
                      {T('common.submit')}
                    </Button>
                  </Box>
                </Grid>
              </Grid>
            </Form>
          )}
        </Formik>
      </Box>
    </Paper>
  )
}

export default Step2
