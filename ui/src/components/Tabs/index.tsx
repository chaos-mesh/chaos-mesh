import { Box, Tab, TabProps, Tabs } from '@material-ui/core'

import React from 'react'

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
  const onChange = (_: React.ChangeEvent<{}>, newValue: number) => setValue(newValue)

  return (
    <>
      <Tabs variant="scrollable" indicatorColor="primary" textColor="primary" value={value} onChange={onChange}>
        {tabs.map(({ label, ...other }: TabProps, index: number) => {
          return <Tab key={index} label={label} {...a11yProps(index)} {...other} />
        })}
      </Tabs>

      <Box flex={1} mt={6} px={3}>
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
    </>
  )
}

export default VerticalTabs
