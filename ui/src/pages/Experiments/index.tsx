import React from 'react'
import { Link } from 'react-router-dom'

import { Box, Button, Card, CardContent, Typography } from '@material-ui/core'
import AddIcon from '@material-ui/icons/Add'

export default function Experiments() {
  return (
    <>
      <Box mb={5}>
        <Button
          variant="contained"
          color="primary"
          className="button-new"
          startIcon={<AddIcon />}
          component={Link}
          to="new-experiment"
        >
          New Experiment
        </Button>
      </Box>

      <Card>
        <CardContent>
          <Typography
            component={Link}
            to="/experiments/tikv-failure"
            variant="h6"
            color="textPrimary"
          >
            tikv-failure
          </Typography>
        </CardContent>
      </Card>
    </>
  )
}
