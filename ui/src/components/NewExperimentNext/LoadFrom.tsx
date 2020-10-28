import { Box, Button } from '@material-ui/core'
import { setAlert, setAlertOpen } from 'slices/globalStatus'

import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import Paper from 'components-mui/Paper'
import PaperTop from 'components/PaperTop'
import React from 'react'
import T from 'components/T'
import { toTitleCase } from 'lib/utils'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'
import { yamlToExperiment } from 'lib/formikhelpers'

interface LoadFromProps {
  from: 'experiments' | 'archives' | 'yaml'
}

const LoadFrom: React.FC<LoadFromProps> = ({ from }) => {
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
        // setInitialValues(y)
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'common.importSuccessfully' }),
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
    <Paper>
      <PaperTop title={T(`newE.loadFrom${toTitleCase(from)}`)} />
      <Box p={6} maxHeight={450} style={{ overflowY: 'scroll' }}>
        {from === 'yaml' && (
          <Button component="label" variant="outlined" size="small" startIcon={<CloudUploadOutlinedIcon />}>
            {T('common.upload')}
            <input type="file" style={{ display: 'none' }} onChange={handleUploadYAML} />
          </Button>
        )}
      </Box>
    </Paper>
  )
}

export default LoadFrom
