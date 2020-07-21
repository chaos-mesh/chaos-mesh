import 'jsoneditor/dist/jsoneditor.css'

import React from 'react'
import _JSONEditor from 'jsoneditor'
import { withStyles } from '@material-ui/core/styles'

const styles = {
  root: {
    width: '100%',
    height: '100%',
    '& .jsoneditor': {
      borderColor: '#172d72',
    },
    '& .jsoneditor-menu': {
      background: '#172d72',
      borderColor: '#172d72',
    },
  },
}

interface JSONEditorProps {
  classes: Record<'root', string>
  json: object | null
}

class JSONEditor extends React.Component<JSONEditorProps> {
  private editorRef = React.createRef<HTMLDivElement>()
  private editor: _JSONEditor | null = null

  componentDidMount() {
    const options = {
      enableSort: false,
      enableTransform: false,
      mode: 'form' as 'form',
      search: false,
    }

    this.editor = new _JSONEditor(this.editorRef.current!, options)
    this.editor.set(this.props.json)
    this.editor?.expandAll()
  }

  componentWillUnmount() {
    this.editor?.destroy()
  }

  render() {
    return <div className={this.props.classes.root} ref={this.editorRef} />
  }
}

export default withStyles(styles)(JSONEditor)
