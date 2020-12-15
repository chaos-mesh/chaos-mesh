import { setAlert, setAlertOpen } from 'slices/globalStatus'
import { setBasic, setKindAction, setTarget } from 'slices/experiments'

import { Button } from '@material-ui/core'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import React from 'react'
import T from 'components/T'
import _snakecase from 'lodash.snakecase'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'
import { yamlToExperiment } from 'lib/formikhelpers'

const YAML = () => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const handleUploadYAML = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files![0]

    const reader = new FileReader()
    reader.onload = function (e) {
      try {
        const y = yamlToExperiment(yaml.safeLoad(e.target!.result as string))
        if (process.env.NODE_ENV === 'development') {
          console.debug('Debug yamlToExperiment:', y)
        }

        const kind = y.target.kind
        dispatch(setKindAction([kind, y.target[_snakecase(kind)].action ?? '']))
        dispatch(setTarget(y.target))
        dispatch(setBasic(y.basic))
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'common.loadSuccessfully' }),
          })
        )
      } catch (e) {
        console.error(e)
        dispatch(
          setAlert({
            type: 'error',
            message: e.message,
          })
        )
      } finally {
        dispatch(setAlertOpen(true))
      }
    }
    reader.readAsText(f)
  }

  return (
    <Button component="label" variant="outlined" size="small" startIcon={<CloudUploadOutlinedIcon />}>
      {T('common.upload')}
      <input type="file" style={{ display: 'none' }} onChange={handleUploadYAML} />
    </Button>
  )
}

export default YAML
