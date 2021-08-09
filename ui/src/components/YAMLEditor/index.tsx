import 'ace-builds'
import 'ace-builds/src-min-noconflict/theme-tomorrow_night_eighties'
import 'ace-builds/src-min-noconflict/mode-yaml'
import 'ace-builds/src-min-noconflict/theme-tomorrow'

import AceEditor, { IAceEditorProps } from 'react-ace'
import { Box, Button } from '@material-ui/core'
import { useStoreDispatch, useStoreSelector } from 'store'

import { Ace } from 'ace-builds'
import CloudDownloadOutlinedIcon from '@material-ui/icons/CloudDownloadOutlined'
import PublishIcon from '@material-ui/icons/Publish'
import Space from 'components-mui/Space'
import T from 'components/T'
import fileDownload from 'js-file-download'
import { memo } from 'react'
import { setConfirm } from 'slices/globalStatus'
import { useIntl } from 'react-intl'
import { useState } from 'react'
import yaml from 'js-yaml'

interface YAMLEditorProps {
  name?: string
  data?: string
  mountEditor?: (editor: Ace.Editor) => void
  onUpdate?: (data: any) => void
  download?: boolean
  aceProps?: IAceEditorProps
}

const YAMLEditor: React.FC<YAMLEditorProps> = ({ name, data, mountEditor, onUpdate, download, aceProps }) => {
  const intl = useIntl()

  const { theme } = useStoreSelector((state) => state.settings)
  const dispatch = useStoreDispatch()

  const [editor, setEditor] = useState<Ace.Editor>()

  const handleOnLoad = (editor: Ace.Editor) => {
    setEditor(editor)

    typeof mountEditor === 'function' && mountEditor(editor)
  }

  const handleSelect = () => {
    dispatch(
      setConfirm({
        title: `${T('common.update', intl)} ${name}`,
        handle: handleOnUpdate,
      })
    )
  }

  const handleOnUpdate = () => {
    typeof onUpdate === 'function' && onUpdate(yaml.load(editor!.getValue()))
  }

  const handleDownloadExperiment = () => fileDownload(editor!.getValue(), `${name}.yaml`)

  return (
    <Box position="relative" width="100%" height="100%">
      <AceEditor
        onLoad={handleOnLoad}
        width="100%"
        height="100%"
        style={{ borderBottomLeftRadius: 4, borderBottomRightRadius: 4 }}
        mode="yaml"
        theme={theme === 'light' ? 'tomorrow' : 'tomorrow_night_eighties'}
        value={data}
        {...aceProps}
      />
      {(typeof onUpdate === 'function' || download) && (
        <Space
          direction="row"
          sx={{ position: 'absolute', top: (theme) => theme.spacing(1.5), right: (theme) => theme.spacing(3) }}
        >
          {download && (
            <Button
              variant="outlined"
              size="small"
              startIcon={<CloudDownloadOutlinedIcon />}
              onClick={handleDownloadExperiment}
            >
              {T('common.download')}
            </Button>
          )}
          {typeof onUpdate === 'function' && (
            <Button variant="outlined" color="primary" size="small" startIcon={<PublishIcon />} onClick={handleSelect}>
              {T('common.update')}
            </Button>
          )}
        </Space>
      )}
    </Box>
  )
}

export default memo(YAMLEditor, (prevProps, nextProps) => {
  if (prevProps.data !== nextProps.data) {
    return false
  }

  return true
})
