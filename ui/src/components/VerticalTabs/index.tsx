import React from 'react'
import { Box, Tabs, Tab, TabProps } from '@material-ui/core'
import { makeStyles, Theme } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) => ({
  tabs: {
    flexShrink: 0,
    borderRight: `1px solid ${theme.palette.divider}`,
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
}

export default function VerticalTabs({ tabs, tabPanels }: VerticalTabsProps) {
  const classes = useStyles()
  const [value, setValue] = React.useState(0)

  const handleChange = (event: React.ChangeEvent<{}>, newValue: number) => {
    setValue(newValue)
  }

  return (
    <Box display="flex" height="100%">
      <Tabs className={classes.tabs} orientation="vertical" variant="scrollable" value={value} onChange={handleChange}>
        {tabs.map(({ label, ...other }: TabProps, index: number) => {
          return <Tab key={index} label={label} {...a11yProps(index)} {...other} />
        })}
      </Tabs>

      {tabPanels.map((panel: React.ReactNode, index: number) => {
        return (
          <Box
            key={index}
            role="tabpanel"
            hidden={value !== index}
            id={`vertical-tabpanel-${index}`}
            aria-labelledby={`vertical-tab-${index}`}
            flexGrow={1}
            p={{ sm: 3, md: 6 }}
          >
            {value === index && panel}
          </Box>
        )
      })}
    </Box>
  )
}
