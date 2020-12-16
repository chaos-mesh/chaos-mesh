import { Box, Card, GridList, GridListTile, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import React, { useRef } from 'react'

import T from 'components/T'
import YAML from 'components/YAML'
import { getStore } from 'lib/idb'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles({
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
  },
})

const Predefined = () => {
  const theme = useTheme()
  const isDesktopScreen = useMediaQuery(theme.breakpoints.down('md'))
  const classes = useStyles()
  const idb = useRef(getStore('predefined'))

  return (
    <Box overflow="hidden">
      <GridList className={classes.gridList} cols={isDesktopScreen ? 1.5 : 3.5} spacing={9} cellHeight="auto">
        <GridListTile>
          <YAML callback={() => {}} />
        </GridListTile>
        <GridListTile>
          <Card className={classes.card} variant="outlined">
            <Box display="flex" justifyContent="center" alignItems="center" height={88}>
              <Box display="flex" justifyContent="center" alignItems="center" flex={1}></Box>
              <Box
                display="flex"
                justifyContent="center"
                alignItems="center"
                flex={2}
                px={1.5}
                textAlign="center"
              ></Box>
            </Box>
          </Card>
        </GridListTile>
      </GridList>
    </Box>
  )
}

export default Predefined
