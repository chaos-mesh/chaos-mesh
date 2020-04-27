import React from 'react'
import { Button } from '@material-ui/core'
import CloudUploadOutlinedIcon from '@material-ui/icons/CloudUploadOutlined'
import PageBar from '../../components/PageBar'
import ToolBar from '../../components/ToolBar'
import Container from '../../components/Container'

export default function NewExperiment() {
  return (
    <>
      <PageBar />
      <ToolBar>
        <Button variant="outlined" startIcon={<CloudUploadOutlinedIcon />}>
          Yaml File
        </Button>
      </ToolBar>

      <Container>Form Step</Container>
    </>
  )
}
