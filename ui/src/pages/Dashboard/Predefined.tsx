import { Box, Card, GridList, GridListTile, Typography } from '@material-ui/core'
import React, { useRef } from 'react'

import AddIcon from '@material-ui/icons/Add'
import T from 'components/T'
import clsx from 'clsx'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => ({
  gridList: {
    flexWrap: 'nowrap',
    // Promote the list into his own layer on Chrome. This cost memory but helps keeping high FPS.
    transform: 'translateZ(0)',
  },
  card: {
    cursor: 'pointer',
  },
  addCard: {
    width: 210,
    borderStyle: 'dashed',
    '&:hover': {
      color: theme.palette.primary.main,
      borderColor: theme.palette.primary.main,
    },
  },
}))

const Predefined = () => {
  const classes = useStyles()
  const dbRef = useRef()

  return (
    <Box overflow="hidden">
      <GridList className={classes.gridList} spacing={9} cellHeight="auto">
        <GridListTile>
          <Card className={clsx(classes.card, classes.addCard)} variant="outlined">
            <Box display="flex" justifyContent="center" alignItems="center" height={88}>
              <Box display="flex" justifyContent="center" alignItems="center" flex={1}>
                <AddIcon />
              </Box>
              <Box display="flex" justifyContent="center" alignItems="center" flex={2} px={1.5} textAlign="center">
                <Typography variant="button">{T('common.add')}</Typography>
              </Box>
            </Box>
          </Card>
        </GridListTile>
      </GridList>
    </Box>
  )
}

export default Predefined
