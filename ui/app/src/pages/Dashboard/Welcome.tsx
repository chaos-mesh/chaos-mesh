/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import AccountTreeOutlinedIcon from '@mui/icons-material/AccountTreeOutlined'
import ScheduleIcon from '@mui/icons-material/Schedule'
import ScienceOutlinedIcon from '@mui/icons-material/ScienceOutlined'
import { Button, Grid } from '@mui/material'
import { makeStyles } from '@mui/styles'
import { useTour } from '@reactour/tour'
import { Link } from 'react-router-dom'

import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import Space from '@ui/mui-extends/esm/Space'

import i18n from 'components/T'

const useStyles = makeStyles((theme) => ({
  space: {
    width: '75%',
    [theme.breakpoints.down('md')]: {
      width: 'unset',
    },
  },
}))

const Welcome = () => {
  const classes = useStyles()

  const { setIsOpen } = useTour()

  return (
    <Grid container spacing={6}>
      <Grid item xs={6}>
        <Paper style={{ height: '100%' }}>
          <Space className={classes.space}>
            <PaperTop title={i18n('dashboard.tutorial.title')} subtitle={i18n('dashboard.tutorial.desc')} />
            <Button
              className="tutorial-end"
              variant="contained"
              color="primary"
              fullWidth
              onClick={() => setIsOpen(true)}
            >
              {i18n('common.tutorial')}
            </Button>
            <PaperTop title={i18n('dashboard.newbie')} subtitle={i18n('dashboard.newbieDesc')} />
            <Button
              className="tutorial-newE"
              component={Link}
              to="/experiments/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<ScienceOutlinedIcon />}
            >
              {i18n('newE.title')}
            </Button>
          </Space>
        </Paper>
      </Grid>
      <Grid item xs={6}>
        <Paper style={{ height: '100%' }}>
          <Space className={classes.space}>
            <PaperTop title={i18n('dashboard.startAWorkflow')} subtitle={i18n('dashboard.startAWorkflowDesc')} />
            <Button
              className="tutorial-newW"
              component={Link}
              to="/workflows/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<AccountTreeOutlinedIcon />}
            >
              {i18n('newW.title')}
            </Button>
            <PaperTop title={i18n('dashboard.startASchedule')} subtitle={i18n('dashboard.startAScheduleDesc')} />
            <Button
              className="tutorial-newS"
              component={Link}
              to="/schedules/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<ScheduleIcon />}
            >
              {i18n('newS.title')}
            </Button>
          </Space>
        </Paper>
      </Grid>
    </Grid>
  )
}

export default Welcome
