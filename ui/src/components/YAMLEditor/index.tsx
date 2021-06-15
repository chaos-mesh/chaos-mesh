import 'ace-builds'
import 'ace-builds/src-min-noconflict/theme-tomorrow_night_eighties'
import 'ace-builds/src-min-noconflict/mode-yaml'
import 'ace-builds/src-min-noconflict/theme-tomorrow'

import AceEditor, { IAceEditorProps } from 'react-ace'

import { Ace } from 'ace-builds'
import React from 'react'
import { useStoreSelector } from 'store'

interface YAMLEditorProps {
  data?: string
  mountEditor?: (editor: Ace.Editor) => void
  aceProps?: IAceEditorProps
}

const YAMLEditor: React.FC<YAMLEditorProps> = ({ data, mountEditor, aceProps }) => {
  const { theme } = useStoreSelector((state) => state.settings)

  const handleOnLoad = (editor: Ace.Editor) => typeof mountEditor === 'function' && mountEditor(editor)

  return (
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
  )
}

export default React.memo(YAMLEditor)
