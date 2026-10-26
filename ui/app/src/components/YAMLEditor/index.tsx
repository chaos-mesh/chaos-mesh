/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import Space from '@/mui-extends/Space'
import { useComponentActions } from '@/zustand/component'
import { useSystemStore } from '@/zustand/system'
import CloudDownloadOutlinedIcon from '@mui/icons-material/CloudDownloadOutlined'
import PublishIcon from '@mui/icons-material/Publish'
import { Box, Button } from '@mui/material'
import { type Editor } from 'ace-builds'
import 'ace-builds/src-noconflict/ace'
import 'ace-builds/src-noconflict/mode-yaml'
import 'ace-builds/src-noconflict/theme-tomorrow'
import 'ace-builds/src-noconflict/theme-tomorrow_night'
import fileDownload from 'js-file-download'
import yaml from 'js-yaml'
import { memo, useState } from 'react'
import AceEditor, { IAceEditorProps } from 'react-ace'
import { useIntl } from 'react-intl'

import i18n from '@/components/T'

interface YAMLEditorProps {
  name?: string
  data?: string
  mountEditor?: (editor: Editor) => void
  onUpdate?: (data: any) => void
  download?: boolean
  aceProps?: IAceEditorProps
}

const YAMLEditor: ReactFCWithChildren<YAMLEditorProps> = ({
  name,
  data,
  mountEditor,
  onUpdate,
  download,
  aceProps,
}) => {
  const intl = useIntl()

  const theme = useSystemStore((state) => state.theme)
  const { setConfirm } = useComponentActions()

  const [editor, setEditor] = useState<Editor>()

  const handleOnLoad = (editor: Editor) => {
    setEditor(editor)

    typeof mountEditor === 'function' && mountEditor(editor)
  }

  const handleSelect = () => {
    setConfirm({
      title: `${i18n('common.update', intl)} ${name}`,
      handle: handleOnUpdate,
    })
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
        theme={theme === 'light' ? 'tomorrow' : 'tomorrow_night'}
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
              {i18n('common.download')}
            </Button>
          )}
          {typeof onUpdate === 'function' && (
            <Button variant="outlined" color="primary" size="small" startIcon={<PublishIcon />} onClick={handleSelect}>
              {i18n('common.update')}
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
