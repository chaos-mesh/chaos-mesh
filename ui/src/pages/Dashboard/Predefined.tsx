import { Box, Card, GridList, GridListTile, Typography, useMediaQuery, useTheme } from '@material-ui/core'
import { PreDefinedValue, getDB } from 'lib/idb'
import React, { useEffect, useRef, useState } from 'react'

import YAML from 'components/YAML'
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
    height: '100%',
  },
})

const Predefined = () => {
  const theme = useTheme()
  const isDesktopScreen = useMediaQuery(theme.breakpoints.down('md'))
  const classes = useStyles()

  const idb = useRef(getDB())

  const [experiments, setExperiments] = useState<PreDefinedValue[]>([])

  async function getExperiments() {
    setExperiments(await (await idb.current).getAll('predefined'))
  }

  useEffect(() => {
    getExperiments()
  }, [])

  const saveExperiment = async (y: any) => {
    const db = await idb.current

    await db.put('predefined', {
      name: y.metadata.name,
      kind: y.kind,
      yaml: y,
    })

    getExperiments()
  }

  return (
    <Box overflow="hidden">
      <GridList className={classes.gridList} cols={isDesktopScreen ? 1.5 : 3.5} spacing={9} cellHeight={88}>
        <GridListTile>
          <YAML callback={saveExperiment} buttonProps={{ className: classes.addCard }} />
        </GridListTile>
        {experiments.map((d) => (
          <GridListTile key={d.name}>
            <Card className={classes.card} variant="outlined">
              <Box display="flex" justifyContent="center" alignItems="center">
                <Box display="flex" justifyContent="center" alignItems="center" flex={1}></Box>
                <Box display="flex" justifyContent="center" alignItems="center" flex={2} px={1.5} textAlign="center">
                  <Typography>{d.name}</Typography>
                </Box>
              </Box>
            </Card>
          </GridListTile>
        ))}
      </GridList>
    </Box>
  )
}

export default Predefined
