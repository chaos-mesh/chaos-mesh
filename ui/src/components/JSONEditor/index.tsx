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
    '& .jsoneditor > .jsoneditor-menu': {
      background: '#172d72',
      borderColor: '#172d72',
    },
  },
}

interface JSONEditorProps {
  classes: Record<'root', string>
  name?: string
  json: object | null
  mountEditor?: (editor: _JSONEditor) => void
}

class JSONEditor extends React.Component<JSONEditorProps> {
  private editorRef = React.createRef<HTMLDivElement>()
  private editor: _JSONEditor | null = null

  componentDidMount() {
    const options = {
      enableSort: false,
      enableTransform: false,
      name: this.props.name,
      mode: 'tree' as 'tree',
      search: false,
    }

    this.editor = new _JSONEditor(this.editorRef.current!, options)
    this.editor.set(this.props.json)
    this.editor.expandAll()

    typeof this.props.mountEditor === 'function' && this.props.mountEditor(this.editor)
  }

  componentWillUnmount() {
    this.editor?.destroy()
  }

  render() {
    return <div className={this.props.classes.root} ref={this.editorRef} />
  }
}

export default withStyles(styles)(JSONEditor)
