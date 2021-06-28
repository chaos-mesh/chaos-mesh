import { Button, Typography } from '@material-ui/core'

import { Ace } from 'ace-builds'
import Paper from 'components-mui/Paper'
import PublishIcon from '@material-ui/icons/Publish'
import Space from 'components-mui/Space'
import T from 'components/T'
import YAML from 'components/YAML'
import loadable from '@loadable/component'
import { setAlert } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useState } from 'react'
import { useStoreDispatch } from 'store'
import yaml from 'js-yaml'

const YAMLEditor = loadable(() => import('components/YAMLEditor'))

interface ByYAMLProps {
  callback?: (data: any) => void
}

const ByYAML: React.FC<ByYAMLProps> = ({ callback }) => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [empty, setEmpty] = useState(true)
  const [yamlEditor, setYAMLEditor] = useState<Ace.Editor>()

  const onChange = (value: string) => setEmpty(value === '')

  const handleUploadYAMLCallback = (y: any) => yamlEditor?.setValue(y)

  const handleSubmit = () => {
    const data = yaml.load(yamlEditor!.getValue())

    callback && callback(data)

    dispatch(
      setAlert({
        type: 'success',
        message: T('confirm.success.load', intl),
      })
    )
  }

  return (
    <Space spacing={6}>
      <Typography variant="body2" color="textSecondary">
        {T('newE.byYAMLDesc')}
      </Typography>
      <Paper sx={{ height: 600, p: 0 }}>
        <YAMLEditor mountEditor={setYAMLEditor} aceProps={{ onChange }} />
      </Paper>
      <Space direction="row" justifyContent="flex-end">
        <YAML callback={handleUploadYAMLCallback} />
        <Button
          variant="contained"
          color="primary"
          startIcon={<PublishIcon />}
          size="small"
          disabled={empty}
          onClick={handleSubmit}
        >
          {T('common.submit')}
        </Button>
      </Space>
    </Space>
  )
}

export default ByYAML
