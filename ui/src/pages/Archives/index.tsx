import React from 'react'

import PageBar from '../../components/PageBar'
import ToolBar from '../../components/ToolBar'
import Container from '../../components/Container'

export default function Archives() {
  return (
    <>
      <PageBar breadcrumbs={[{ name: ' Archives' }]} />
      <ToolBar />

      <Container maxWidth="xl">Archived Experiment List</Container>
    </>
  )
}
