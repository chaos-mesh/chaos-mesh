import React from 'react'

import PageBar from '../../components/PageBar'
import ToolBar from '../../components/ToolBar'
import Container from '../../components/Container'

export default function Events() {
  return (
    <>
      <PageBar breadcrumbs={[{ name: ' Events' }]} />
      <ToolBar />

      <Container maxWidth="xl">Event List</Container>
    </>
  )
}
