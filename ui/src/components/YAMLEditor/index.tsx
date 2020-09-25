import 'ace-builds'
import 'ace-builds/webpack-resolver'
import 'ace-builds/src-noconflict/mode-yaml'
import 'ace-builds/src-noconflict/theme-tomorrow'
import 'ace-builds/src-noconflict/theme-tomorrow_night_eighties'

import React, { useEffect, useRef } from 'react'

import AceEditor from 'react-ace'
import { Theme } from 'slices/settings'
import ace from 'ace-builds'

interface YAMLEditorProps {
  theme: Theme
  data: object | null
  mountEditor?: (editor: ace.Ace.Editor) => void
}

const YAMLEditor: React.FC<YAMLEditorProps> = ({ theme, mountEditor }) => {
  const innerEditor = useRef(null)

  useEffect(() => {
    typeof mountEditor === 'function' && mountEditor(innerEditor.current!)

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <AceEditor
      ref={innerEditor}
      width="100%"
      height="100%"
      mode="yaml"
      theme={theme === 'light' ? 'tomorrow' : 'tomorrow_night_eighties'}
    />
  )
}

export default YAMLEditor
