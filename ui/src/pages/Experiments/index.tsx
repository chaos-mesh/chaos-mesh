import React, { useState } from 'react'
import { Link } from 'react-router-dom'

import { Button, Card, CardContent, Drawer, Typography } from '@material-ui/core'
import AddIcon from '@material-ui/icons/Add'

import PageBar from '../../components/PageBar'
import ToolBar from '../../components/ToolBar'
import Container from '../../components/Container'
import NewExperiment from '../../pages/Experiments/New'

export default function Experiments() {
  const [isOpen, setIsOpen] = useState(false)

  const toggleDrawer = (isOpen: boolean) => () => {
    setIsOpen(isOpen)
  }

  // FIXME: console warning: findDOMNode is deprecated in StrictMode.
  // https://github.com/mui-org/material-ui/issues/13394
  return (
    <>
      <PageBar />
      <ToolBar>
        <Button variant="outlined" startIcon={<AddIcon />} onClick={toggleDrawer(true)}>
          New Experiment
        </Button>
      </ToolBar>

      <Container>
        <Card>
          <CardContent>
            <Typography component={Link} to="/experiments/tikv-failure" variant="h6" color="primary">
              tikv-failure
            </Typography>
          </CardContent>
        </Card>
      </Container>

      {/* New Experiment Stepper Drawer */}
      <Drawer anchor="right" open={isOpen} onClose={toggleDrawer(false)}>
        <NewExperiment />
      </Drawer>
    </>
  )
}
