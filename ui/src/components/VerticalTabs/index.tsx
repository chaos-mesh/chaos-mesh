import { Box, Tab, TabProps, Tabs } from '@material-ui/core'
import { Theme, makeStyles } from '@material-ui/core/styles'

import React from 'react'

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    display: 'flex',
  },
  tabs: {
    borderRight: `1px solid ${theme.palette.divider}`,
    '& .MuiTabs-indicator': {
      backgroundColor: theme.palette.primary.main,
    },
  },
  main: {
    flex: 1,
    paddingLeft: theme.spacing(6),
    paddingRight: theme.spacing(3),
  },
}))

function a11yProps(index: number) {
  return {
    id: `vertical-tab-${index}`,
    'aria-controls': `vertical-tabpanel-${index}`,
  }
}

interface VerticalTabsProps {
  tabs: TabProps[]
  tabPanels: React.ReactNode[]
  tabIndex: number
  setTabIndex: (index: number) => void
}

const VerticalTabs: React.FC<VerticalTabsProps> = ({ tabs, tabPanels, tabIndex: value, setTabIndex: setValue }) => {
  const classes = useStyles()

  const onChange = (_: React.ChangeEvent<{}>, newValue: number) => setValue(newValue)

  return (
    <Box className={classes.root}>
      <Tabs className={classes.tabs} orientation="vertical" value={value} onChange={onChange}>
        {tabs.map(({ label, ...other }: TabProps, index: number) => {
          return <Tab key={index} label={label} {...a11yProps(index)} {...other} />
        })}
      </Tabs>

      <Box className={classes.main}>
        {tabPanels.map((panel: React.ReactNode, index: number) => {
          return (
            <Box
              key={index}
              role="tabpanel"
              id={`vertical-tabpanel-${index}`}
              aria-labelledby={`vertical-tab-${index}`}
              hidden={value !== index}
            >
              {value === index && panel}
            </Box>
          )
        })}
      </Box>
    </Box>
  )
}

export default VerticalTabs
