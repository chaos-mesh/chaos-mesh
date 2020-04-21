import React from 'react'

import PageBar from '../../components/PageBar'
import ToolBar from '../../components/ToolBar'
import Container from '../../components/Container'

export default function Overview() {
  return (
    <>
      <PageBar breadcrumbs={[{ name: 'Overview' }]} />
      <ToolBar />

      <Container>Metrics Charts</Container>
    </>
  )
}
