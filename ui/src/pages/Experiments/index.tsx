import React from 'react'
import { Link } from 'react-router-dom'

import { Button, Card, CardContent, Typography } from '@material-ui/core'
import AddIcon from '@material-ui/icons/Add'
import PageBar from '../../components/PageBar'
import ToolBar from '../../components/ToolBar'
import Container from '../../components/Container'

export default function Experiments() {
  return (
    <>
      <PageBar />
      <ToolBar>
        <Button variant="outlined" startIcon={<AddIcon />} component={Link} to="new-experiment">
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
    </>
  )
}
