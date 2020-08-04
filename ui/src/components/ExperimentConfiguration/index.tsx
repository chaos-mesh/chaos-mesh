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
import { ChaosKindKeyMap } from 'lib/formikhelpers'
import { ExperimentDetail } from 'api/experiments.type'
import React from 'react'
import { format } from 'lib/dayjs'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

interface ExperimentConfigurationProps {
  experimentDetail: ExperimentDetail | ArchiveDetail
}

const ExperimentConfiguration: React.FC<ExperimentConfigurationProps> = ({ experimentDetail: e }) => (
  <Grid container>
    <Grid item md={4}>
      <Box mt={3} ml="16px">
        <Typography variant="subtitle2" gutterBottom>
          Basic
        </Typography>
      </Box>

      <Table size="small">
        <TableBody>
          <TableRow>
            <TableCell>Name</TableCell>
            <TableCell>
              <Typography variant="body2" color="textSecondary">
                {e.name}
              </Typography>
            </TableCell>
          </TableRow>

          <TableRow>
            <TableCell>Kind</TableCell>
            <TableCell>
              <Typography variant="body2" color="textSecondary">
                {e.kind}
              </Typography>
            </TableCell>
          </TableRow>

          {['PodChaos', 'NetworkChaos', 'IoChaos'].includes(e.kind) && (
            <TableRow>
              <TableCell>Action</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {(e.experiment_info.target[ChaosKindKeyMap[e.kind].key] as any).action}
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
          Meta
        </Typography>
      </Box>

      <Table size="small">
        <TableBody>
          <TableRow>
            <TableCell>Namespace</TableCell>
            <TableCell>
              <Typography variant="body2" color="textSecondary">
                {e.namespace}
              </Typography>
            </TableCell>
          </TableRow>

          <TableRow>
            <TableCell>UUID</TableCell>
            <TableCell>
              <Typography variant="body2" color="textSecondary">
                {e.uid}
              </Typography>
            </TableCell>
          </TableRow>

          {(e as ExperimentDetail).created && (
            <TableRow>
              <TableCell>Created</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {format((e as ExperimentDetail).created)}
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
          Scheduler
        </Typography>
      </Box>

      <Table size="small">
        <TableBody>
          {e.experiment_info.scheduler.cron ? (
            <>
              <TableRow>
                <TableCell>Cron</TableCell>
                <TableCell>
                  <Typography variant="body2" color="textSecondary">
                    {e.experiment_info.scheduler.cron}
                  </Typography>
                </TableCell>
              </TableRow>

              <TableRow>
                <TableCell>Duration</TableCell>
                <TableCell>
                  <Typography variant="body2" color="textSecondary">
                    {e.experiment_info.scheduler.duration ? e.experiment_info.scheduler.duration : 'immediate'}
                  </Typography>
                </TableCell>
              </TableRow>
            </>
          ) : (
            <TableRow>
              <TableCell>Immediate Job</TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </Grid>
  </Grid>
)

export default ExperimentConfiguration
