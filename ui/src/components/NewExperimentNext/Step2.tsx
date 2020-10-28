import { Box, Button, Grid, MenuItem } from '@material-ui/core'
import { Form, Formik } from 'formik'
import { LabelField, SelectField, TextField } from 'components/FormField'
import React, { useEffect } from 'react'
import { RootState, useStoreDispatch } from 'store'
import { createStyles, makeStyles } from '@material-ui/core/styles'
import { setBasic, setStep2 } from 'slices/experiments'

import AdvancedOptions from 'components/AdvancedOptions'
import CheckCircleOutlineIcon from '@material-ui/icons/CheckCircleOutline'
import Paper from 'components-mui/Paper'
import PaperTop from 'components/PaperTop'
import PublishIcon from '@material-ui/icons/Publish'
import Scheduler from './form/Scheduler'
import Scope from './form/Scope'
import SkeletonN from 'components/SkeletonN'
import T from 'components/T'
import UndoIcon from '@material-ui/icons/Undo'
import basicData from './data/basic'
import { getNamespaces } from 'slices/experiments'
import { useSelector } from 'react-redux'

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

  const { namespaces, step2 } = useSelector((state: RootState) => state.experiments)
  const dispatch = useStoreDispatch()

  useEffect(() => {
    dispatch(getNamespaces())
  }, [dispatch])

  const handleOnSubmitStep2 = (values: Record<string, any>) => {
    if (process.env.NODE_ENV === 'development') {
      console.debug('Debug handleSubmitStep2', values)
    }

    dispatch(setBasic(values))
    dispatch(setStep2(true))
  }

  const handleUndo = () => {
    dispatch(setStep2(false))
  }

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
        <Formik initialValues={basicData} onSubmit={handleOnSubmitStep2}>
          <Form>
            <Grid container spacing={9}>
              <Grid item xs={12} md={6}>
                {namespaces.length ? <Scope namespaces={namespaces} /> : <SkeletonN n={6} />}
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField id="name" name="name" label={T('newE.basic.name')} helperText={T('newE.basic.nameHelper')} />

                {namespaces.length ? (
                  <SelectField
                    id="namespace"
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
                ) : (
                  <SkeletonN n={3} />
                )}

                <AdvancedOptions>
                  <LabelField id="labels" name="labels" label={T('k8s.labels')} isKV />
                  <LabelField id="annotations" name="annotations" label={T('k8s.annotations')} isKV />
                </AdvancedOptions>
                <Scheduler />
                <Box mt={6} textAlign="right">
                  <Button type="submit" variant="contained" color="primary" startIcon={<PublishIcon />}>
                    {T('common.submit')}
                  </Button>
                </Box>
              </Grid>
            </Grid>
          </Form>
        </Formik>
      </Box>
    </Paper>
  )
}

export default Step2
