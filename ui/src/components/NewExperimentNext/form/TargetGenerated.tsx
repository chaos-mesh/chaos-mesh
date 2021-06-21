import { AutocompleteMultipleField, LabelField, SelectField, Submit, TextField } from 'components/FormField'
import { Form, Formik, FormikErrors, FormikTouched, getIn } from 'formik'
import { Kind, Spec } from '../data/target'
import { useEffect, useState } from 'react'
import { useStoreDispatch, useStoreSelector } from 'store'

import AdvancedOptions from 'components/AdvancedOptions'
import { MenuItem } from '@material-ui/core'
import { ObjectSchema } from 'yup'
import Scope from './Scope'
import Space from 'components-mui/Space'
import T from 'components/T'
import _snakecase from 'lodash.snakecase'
import basicData from '../data/basic'
import { clearNetworkTargetPods } from 'slices/experiments'

interface TargetGeneratedProps {
  kind?: Kind | ''
  data: Spec
  validationSchema: ObjectSchema
  onSubmit: (values: Record<string, any>) => void
}

const TargetGenerated: React.FC<TargetGeneratedProps> = ({ kind, data, validationSchema, onSubmit }) => {
  const { namespaces, target } = useStoreSelector((state) => state.experiments)
  const dispatch = useStoreDispatch()

  let initialValues = Object.entries(data).reduce((acc, [k, v]) => {
    if (v instanceof Object && v.field) {
      acc[k] = v.value
    } else {
      acc[k] = v
    }

    return acc
  }, {} as Record<string, any>)

  if (kind === 'NetworkChaos') {
    const action = initialValues.action
    delete initialValues.action
    const direction = initialValues.direction
    delete initialValues.direction
    const externalTargets = initialValues.external_targets
    delete initialValues.external_targets

    initialValues = {
      action,
      [action]: initialValues,
      direction,
      external_targets: externalTargets,
    }
  }

  const [init, setInit] = useState(initialValues)

  useEffect(() => {
    if (target['kind']) {
      setInit({
        ...initialValues,
        ...target[_snakecase(kind)],
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [target])

  const parseDataToFormFields = (
    errors: FormikErrors<Record<string, any>>,
    touched: FormikTouched<Record<string, any>>
  ) => {
    const rendered = Object.entries(data)
      .filter(([_, v]) => v && v instanceof Object && v.field)
      .map(([k, v]) => {
        if (kind === 'NetworkChaos' && k !== 'direction' && k !== 'external_targets') {
          k = `${data.action}.${k}`
        }

        switch (v.field) {
          case 'text':
            return (
              <TextField
                key={k}
                name={k}
                label={v.label}
                helperText={getIn(touched, k) && getIn(errors, k) ? getIn(errors, k) : v.helperText}
                error={getIn(touched, k) && getIn(errors, k) ? true : false}
                {...v.inputProps}
              />
            )
          case 'number':
            return (
              <TextField
                key={k}
                type="number"
                name={k}
                label={v.label}
                helperText={getIn(touched, k) && getIn(errors, k) ? getIn(errors, k) : v.helperText}
                error={getIn(errors, k) && getIn(touched, k) ? true : false}
                {...v.inputProps}
              />
            )
          case 'select':
            return (
              <SelectField
                key={k}
                name={k}
                label={v.label}
                helperText={getIn(touched, k) && getIn(errors, k) ? getIn(errors, k) : v.helperText}
                error={getIn(errors, k) && getIn(touched, k) ? true : false}
              >
                {v.items!.map((option: string) => (
                  <MenuItem key={option} value={option}>
                    {option}
                  </MenuItem>
                ))}
              </SelectField>
            )
          case 'label':
            return (
              <LabelField
                key={k}
                name={k}
                label={v.label}
                helperText={v.helperText}
                isKV={v.isKV}
                errorText={getIn(errors, k) && getIn(touched, k) ? getIn(errors, k) : ''}
              />
            )
          case 'autocomplete':
            return (
              <AutocompleteMultipleField
                key={k}
                name={k}
                label={v.label}
                helperText={v.helperText}
                options={v.items!}
              />
            )
          default:
            return null
        }
      })
      .filter((d) => d)

    return <>{rendered.map((d) => d)}</>
  }

  return (
    <Formik enableReinitialize initialValues={init} validationSchema={validationSchema} onSubmit={onSubmit}>
      {({ values, setFieldValue, errors, touched }) => {
        const beforeTargetOpen = () => {
          if (!getIn(values, 'target_scope')) {
            setFieldValue('target_scope', basicData.scope)
          }
        }

        const afterTargetClose = () => {
          if (getIn(values, 'target_scope')) {
            setFieldValue('target_scope', undefined)
            dispatch(clearNetworkTargetPods())
          }
        }

        return (
          <Form>
            <Space>{parseDataToFormFields(errors, touched)}</Space>
            {kind === 'NetworkChaos' && (
              <AdvancedOptions
                title={T('newE.target.network.target.title')}
                beforeOpen={beforeTargetOpen}
                afterClose={afterTargetClose}
              >
                {values.target_scope && (
                  <Scope
                    namespaces={namespaces}
                    scope="target_scope"
                    podsPreviewTitle={T('newE.target.network.target.podsPreview')}
                    podsPreviewDesc={T('newE.target.network.target.podsPreviewHelper')}
                  />
                )}
              </AdvancedOptions>
            )}
            <Submit />
          </Form>
        )
      }}
    </Formik>
  )
}

export default TargetGenerated
