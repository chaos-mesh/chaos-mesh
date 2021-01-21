import 'ace-builds'
import 'ace-builds/src-min-noconflict/theme-tomorrow_night_eighties'
import 'ace-builds/src-min-noconflict/mode-yaml'
import 'ace-builds/src-min-noconflict/theme-tomorrow'

import { Ace } from 'ace-builds'
import AceEditor from 'react-ace'
import React from 'react'
import { Theme } from 'slices/settings'

interface YAMLEditorProps {
  theme: Theme
  data: string
  mountEditor?: (editor: Ace.Editor) => void
}

const YAMLEditor: React.FC<YAMLEditorProps> = ({ theme, data, mountEditor }) => {
  const handleOnLoad = (editor: Ace.Editor) => typeof mountEditor === 'function' && mountEditor(editor)

  return (
    <AceEditor
      onLoad={handleOnLoad}
      width="100%"
      height="100%"
      mode="yaml"
      theme={theme === 'light' ? 'tomorrow' : 'tomorrow_night_eighties'}
      value={data}
    />
  )
}

export default YAMLEditor
