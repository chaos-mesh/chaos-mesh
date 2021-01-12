import {
  Box,
  Grid,
  TableCell as MUITableCell,
  Table,
  TableBody,
  TableRow,
  Typography,
  withStyles,
} from '@material-ui/core'

import { ArchiveDetail } from 'api/archives.type'
import { ExperimentDetail } from 'api/experiments.type'
import React from 'react'
import { RootState } from 'store'
import T from 'components/T'
import { format } from 'lib/dayjs'
import { useSelector } from 'react-redux'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

interface ExperimentConfigurationProps {
  experimentDetail: ExperimentDetail | ArchiveDetail
}

const ExperimentConfiguration: React.FC<ExperimentConfigurationProps> = ({ experimentDetail: e }) => {
  const { lang } = useSelector((state: RootState) => state.settings)

  return (
    <Grid container>
      <Grid item md={4}>
        <Box mt={3} ml="16px">
          <Typography variant="subtitle2" gutterBottom>
            {T('newE.steps.basic')}
          </Typography>
        </Box>

        <Table size="small">
          <TableBody>
            <TableRow>
              <TableCell>{T('newE.basic.name')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {e.name}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('newE.target.kind')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {e.kind}
                </Typography>
              </TableCell>
            </TableRow>

            {['PodChaos', 'NetworkChaos', 'IoChaos'].includes(e.kind) && (
              <TableRow>
                <TableCell>{T('newE.target.action')}</TableCell>
                <TableCell>
                  <Typography variant="body2" color="textSecondary">
                    {e.yaml.spec.action}
                  </Typography>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </Grid>

      <Grid item md={4}>
        <Box mt={3} ml="16px">
          <Typography variant="subtitle2" gutterBottom>
            {T('common.meta')}
          </Typography>
        </Box>

        <Table size="small">
          <TableBody>
            <TableRow>
              <TableCell>{T('newE.basic.namespace')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {e.namespace}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('common.uuid')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {e.uid}
                </Typography>
              </TableCell>
            </TableRow>

            {(e as ExperimentDetail).created && (
              <TableRow>
                <TableCell>{T('experiments.createdAt')}</TableCell>
                <TableCell>
                  <Typography variant="body2" color="textSecondary">
                    {format((e as ExperimentDetail).created, lang)}
                  </Typography>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </Grid>

      <Grid item md={4}>
        <Box mt={3} ml="16px">
          <Typography variant="subtitle2" gutterBottom>
            {T('newE.steps.schedule')}
          </Typography>
        </Box>

        <Table size="small">
          <TableBody>
            {e.yaml.spec.scheduler?.cron ? (
              <>
                <TableRow>
                  <TableCell>Cron</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {e.yaml.spec.scheduler.cron}
                    </Typography>
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>{T('newE.schedule.duration')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {e.yaml.spec.duration || 'immediate'}
                    </Typography>
                  </TableCell>
                </TableRow>
              </>
            ) : (
              <TableRow>
                <TableCell>{T('newE.schedule.continuous')}</TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </Grid>
    </Grid>
  )
}

export default ExperimentConfiguration
