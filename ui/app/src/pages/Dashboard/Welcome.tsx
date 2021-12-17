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

import { Button, Grid } from '@mui/material'

import AccountTreeOutlinedIcon from '@mui/icons-material/AccountTreeOutlined'
import ExperimentIcon from '@ui/mui-extends/esm/Icons/Experiment'
import { Link } from 'react-router-dom'
import Paper from '@ui/mui-extends/esm/Paper'
import PaperTop from '@ui/mui-extends/esm/PaperTop'
import ScheduleIcon from '@mui/icons-material/Schedule'
import Space from '@ui/mui-extends/esm/Space'
import T from 'components/T'
import { makeStyles } from '@mui/styles'
import { useTour } from '@reactour/tour'

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
            <PaperTop title={T('dashboard.tutorial.title')} subtitle={T('dashboard.tutorial.desc')} />
            <Button
              className="tutorial-end"
              variant="contained"
              color="primary"
              fullWidth
              onClick={() => setIsOpen(true)}
            >
              {T('common.tutorial')}
            </Button>
            <PaperTop title={T('dashboard.newbie')} subtitle={T('dashboard.newbieDesc')} />
            <Button
              className="tutorial-newE"
              component={Link}
              to="/experiments/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<ExperimentIcon />}
            >
              {T('newE.title')}
            </Button>
          </Space>
        </Paper>
      </Grid>
      <Grid item xs={6}>
        <Paper style={{ height: '100%' }}>
          <Space className={classes.space}>
            <PaperTop title={T('dashboard.startAWorkflow')} subtitle={T('dashboard.startAWorkflowDesc')} />
            <Button
              className="tutorial-newW"
              component={Link}
              to="/workflows/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<AccountTreeOutlinedIcon />}
            >
              {T('newW.title')}
            </Button>
            <PaperTop title={T('dashboard.startASchedule')} subtitle={T('dashboard.startAScheduleDesc')} />
            <Button
              className="tutorial-newS"
              component={Link}
              to="/schedules/new"
              variant="contained"
              color="primary"
              fullWidth
              startIcon={<ScheduleIcon />}
            >
              {T('newS.title')}
            </Button>
          </Space>
        </Paper>
      </Grid>
    </Grid>
  )
}

export default Welcome
