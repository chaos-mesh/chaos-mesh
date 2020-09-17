import { FormikCtx, StepperFormTargetProps } from 'components/NewExperiment/types'
import { InputAdornment, MenuItem } from '@material-ui/core'
import React, { useEffect } from 'react'
import { SelectField, TextField } from 'components/FormField'

import AdvancedOptions from 'components/AdvancedOptions'
import { RootState } from 'store'
import ScopeStep from '../Scope'
import T from 'components/T'
import { defaultExperimentSchema } from 'components/NewExperiment/constants'
import { getIn } from 'formik'
import { toTitleCase } from 'lib/utils'
import { useFormikContext } from 'formik'
import { useSelector } from 'react-redux'

const actions = ['partition', 'loss', 'delay', 'duplicate', 'corrupt', 'bandwidth']
const direction = ['from', 'to', 'both']

export default function Network(props: StepperFormTargetProps) {
  const { errors, touched, values, setFieldValue }: FormikCtx = useFormikContext()
  const { handleActionChange } = props

  const { namespaces } = useSelector((state: RootState) => state.experiments)

  const initTarget = () => setFieldValue('target.network_chaos.target', defaultExperimentSchema.scope)
  const initPartitionTarget = () => {
    const target = getIn(values, 'target.network_chaos.target')

    setFieldValue(
      'target.network_chaos.target',
      Object.assign(
        {
          ...defaultExperimentSchema.scope,
          mode: 'all',
        },
        target
      )
    )
  }
  const beforeTargetOpen = initTarget
  const afterTargetClose = () => setFieldValue('target.network_chaos.target', undefined)

  // Special operations for partition
  useEffect(() => {
    if (values.target.network_chaos.action === 'partition') {
      initPartitionTarget()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [values.target.network_chaos.action])

  return (
    <>
      <SelectField
        id="target.network_chaos.action"
        name="target.network_chaos.action"
        label={T('newE.target.action')}
        helperText={T('newE.target.network.actionHelper')}
        onChange={handleActionChange}
        onBlur={() => {}} // Delay the form validation with an empty func. If don’t do this, errors will appear early
      >
        {actions.map((option) => (
          <MenuItem key={option} value={option}>
            {toTitleCase(option)}
          </MenuItem>
        ))}
      </SelectField>

      {values.target.network_chaos.action === 'partition' && (
        <SelectField
          id="target.network_chaos.direction"
          name="target.network_chaos.direction"
          label={T('newE.target.network.partition.direction')}
          helperText={T('newE.target.network.partition.directionHelper')}
          error={errors.target?.network_chaos?.direction && touched.target?.network_chaos?.direction ? true : false}
        >
          {direction.map((option) => (
            <MenuItem key={option} value={option}>
              {toTitleCase(option)}
            </MenuItem>
          ))}
        </SelectField>
      )}

      {values.target.network_chaos.action === 'bandwidth' && (
        <>
          <TextField
            id="target.network_chaos.bandwidth.rate"
            name="target.network_chaos.bandwidth.rate"
            label={T('newE.target.network.bandwidth.rate')}
            helperText={T('newE.target.network.bandwidth.rateHelper')}
            error={
              errors.target?.network_chaos?.bandwidth?.rate && touched.target?.network_chaos?.bandwidth?.rate
                ? true
                : false
            }
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.buffer"
            name="target.network_chaos.bandwidth.buffer"
            label={T('newE.target.network.bandwidth.buffer')}
            helperText={T('newE.target.network.bandwidth.bufferHelper')}
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.limit"
            name="target.network_chaos.bandwidth.limit"
            label={T('newE.target.network.bandwidth.limit')}
            helperText={T('newE.target.network.bandwidth.limitHelper')}
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.peakrate"
            name="target.network_chaos.bandwidth.peakrate"
            label={T('newE.target.network.bandwidth.peakrate')}
            helperText={T('newE.target.network.bandwidth.peakrateHelper')}
          />
          <TextField
            type="number"
            id="target.network_chaos.bandwidth.minburst"
            name="target.network_chaos.bandwidth.minburst"
            label={T('newE.target.network.bandwidth.minburst')}
            helperText={T('newE.target.network.bandwidth.minburstHelper')}
          />
        </>
      )}

      {values.target.network_chaos.action === 'corrupt' && (
        <>
          <TextField
            id="target.network_chaos.corrupt.corrupt"
            name="target.network_chaos.corrupt.corrupt"
            label="Corrupt"
            helperText="The percentage of packet corruption"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
            error={
              errors.target?.network_chaos?.corrupt?.corrupt && touched.target?.network_chaos?.corrupt?.corrupt
                ? true
                : false
            }
          />
          <TextField
            id="target.network_chaos.corrupt.correlation"
            name="target.network_chaos.corrupt.correlation"
            label="Correlation"
            helperText="The correlation of corrupt"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
            error={
              errors.target?.network_chaos?.corrupt?.correlation && touched.target?.network_chaos?.corrupt?.correlation
                ? true
                : false
            }
          />
        </>
      )}

      {values.target.network_chaos.action === 'delay' && (
        <>
          <TextField
            id="target.network_chaos.delay.latency"
            name="target.network_chaos.delay.latency"
            label="Latency"
            helperText="The latency of delay"
            autoFocus
            error={
              errors.target?.network_chaos?.delay?.latency && touched.target?.network_chaos?.delay?.latency
                ? true
                : false
            }
          />
          <AdvancedOptions>
            <TextField
              id="target.network_chaos.delay.jitter"
              name="target.network_chaos.delay.jitter"
              label="Jitter"
              helperText="The jitter of delay"
            />
            <TextField
              id="target.network_chaos.delay.correlation"
              name="target.network_chaos.delay.correlation"
              label="Correlation"
              helperText="The correlation of delay"
              InputProps={{
                endAdornment: <InputAdornment position="end">%</InputAdornment>,
              }}
            />
          </AdvancedOptions>
        </>
      )}

      {values.target.network_chaos.action === 'duplicate' && (
        <>
          <TextField
            id="target.network_chaos.duplicate.duplicate"
            name="target.network_chaos.duplicate.duplicate"
            label="Duplicate"
            helperText="The percentage of packet duplication"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
            error={
              errors.target?.network_chaos?.duplicate?.duplicate && touched.target?.network_chaos?.duplicate?.duplicate
                ? true
                : false
            }
          />
          <TextField
            id="target.network_chaos.duplicate.correlation"
            name="target.network_chaos.duplicate.correlation"
            label="Correlation"
            helperText="The correlation of duplicate"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
            error={
              errors.target?.network_chaos?.duplicate?.correlation &&
              touched.target?.network_chaos?.duplicate?.correlation
                ? true
                : false
            }
          />
        </>
      )}

      {values.target.network_chaos.action === 'loss' && (
        <>
          <TextField
            id="target.network_chaos.loss.loss"
            name="target.network_chaos.loss.loss"
            label="Loss"
            helperText="The percentage of packet loss"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
            error={errors.target?.network_chaos?.loss?.loss && touched.target?.network_chaos?.loss?.loss ? true : false}
          />
          <TextField
            id="target.network_chaos.loss.correlation"
            name="target.network_chaos.loss.correlation"
            label="Correlation"
            helperText="The correlation of loss"
            InputProps={{
              endAdornment: <InputAdornment position="end">%</InputAdornment>,
            }}
            error={
              errors.target?.network_chaos?.loss?.correlation && touched.target?.network_chaos?.loss?.correlation
                ? true
                : false
            }
          />
        </>
      )}

      {values.target.network_chaos.action !== '' && (
        <AdvancedOptions
          title={T('newE.target.network.target.title')}
          isOpen={values.target.network_chaos.action === 'partition' ? true : false}
          beforeOpen={beforeTargetOpen}
          afterClose={afterTargetClose}
        >
          {values.target.network_chaos.target && values.target.network_chaos.target.mode && (
            <ScopeStep
              namespaces={namespaces}
              scope="target.network_chaos.target"
              podsPreviewTitle={T('newE.target.network.target.podsPreview')}
              podsPreviewDesc={T('newE.target.network.target.podsPreviewHelper')}
            />
          )}
        </AdvancedOptions>
      )}
    </>
  )
}
