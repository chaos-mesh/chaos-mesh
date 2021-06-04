import { Grid, TableCell as MUITableCell, Table, TableBody, TableRow, Typography, withStyles } from '@material-ui/core'

import { ArchiveDetail } from 'api/archives.type'
import { ExperimentDetail } from 'api/experiments.type'
import React from 'react'
import { RootState } from 'store'
import T from 'components/T'
import { format } from 'lib/luxon'
import { toTitleCase } from 'lib/utils'
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

  const action: string = e.kube_object.spec.action

  return (
    <Grid container>
      <Grid item md={4}>
        <Typography variant="subtitle2" gutterBottom>
          {T('newE.steps.basic')}
        </Typography>

        <Table size="small">
          <TableBody>
            <TableRow>
              <TableCell>{T('common.name')}</TableCell>
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

            {['PodChaos', 'NetworkChaos', 'IOChaos'].includes(e.kind) && (
              <TableRow>
                <TableCell>{T('newE.target.action')}</TableCell>
                <TableCell>
                  <Typography variant="body2" color="textSecondary">
                    {action.includes('-')
                      ? (function () {
                          const split = action.split('-')

                          return toTitleCase(split[0]) + ' ' + toTitleCase(split[1])
                        })()
                      : toTitleCase(action)}
                  </Typography>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </Grid>

      <Grid item md={4}>
        <Typography variant="subtitle2" gutterBottom>
          {T('common.meta')}
        </Typography>

        <Table size="small">
          <TableBody>
            <TableRow>
              <TableCell>{T('common.uuid')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {e.uid}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('k8s.namespace')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {e.namespace}
                </Typography>
              </TableCell>
            </TableRow>

            {(e as ExperimentDetail).created && (
              <TableRow>
                <TableCell>{T('table.created')}</TableCell>
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
        <Typography variant="subtitle2" gutterBottom>
          {T('newE.steps.schedule')}
        </Typography>

        <Table size="small">
          <TableBody>
            {e.kube_object.spec.scheduler?.cron ? (
              <>
                <TableRow>
                  <TableCell>Cron</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {e.kube_object.spec.scheduler.cron}
                    </Typography>
                  </TableCell>
                </TableRow>

                {e.kube_object.spec.duration && (
                  <TableRow>
                    <TableCell>{T('newE.schedule.duration')}</TableCell>
                    <TableCell>
                      <Typography variant="body2" color="textSecondary">
                        {e.kube_object.spec.duration}
                      </Typography>
                    </TableCell>
                  </TableRow>
                )}
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
