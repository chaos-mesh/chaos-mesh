import { Button, ButtonProps } from '@material-ui/core'

import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import React from 'react'
import T from 'components/T'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'

interface YAMLProps {
  callback: (y: any) => void
  buttonProps?: ButtonProps<'label'>
}

const YAML: React.FC<YAMLProps> = ({ callback, buttonProps }) => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const handleUploadYAML = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files![0]

    const reader = new FileReader()
    reader.onload = function (e) {
      try {
        const y = yaml.load(e.target!.result as string)
        if (process.env.NODE_ENV === 'development') {
          console.debug('Debug yamlToExperiment:', y)
        }

        callback(y)

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
      }
    }
    reader.readAsText(f)
  }

  return (
    <Button {...buttonProps} component="label" variant="outlined" size="small" startIcon={<CloudUploadOutlinedIcon />}>
      {T('common.upload')}
      <input type="file" hidden onChange={handleUploadYAML} />
    </Button>
  )
}

export default YAML
