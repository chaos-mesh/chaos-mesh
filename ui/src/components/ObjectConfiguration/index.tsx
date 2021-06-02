import { Grid, TableCell as MUITableCell, Table, TableBody, TableRow, Typography } from '@material-ui/core'

import { ArchiveDetail } from 'api/archives.type'
import { ExperimentDetail } from 'api/experiments.type'
import React from 'react'
import T from 'components/T'
import { format } from 'lib/luxon'
import { toTitleCase } from 'lib/utils'
import { useStoreSelector } from 'store'
import { withStyles } from '@material-ui/styles'

const TableCell = withStyles({
  root: {
    borderBottom: 'none',
  },
})(MUITableCell)

interface ObjectConfigurationProps {
  config: ExperimentDetail | ArchiveDetail
}

const ObjectConfiguration: React.FC<ObjectConfigurationProps> = ({ config }) => {
  const { lang } = useStoreSelector((state) => state.settings)

  const action: string = config.kube_object.spec.action

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
                  {config.name}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('newE.target.kind')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {config.kind}
                </Typography>
              </TableCell>
            </TableRow>

            {['PodChaos', 'NetworkChaos', 'IoChaos'].includes(config.kind) && (
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
                  {config.uid}
                </Typography>
              </TableCell>
            </TableRow>

            <TableRow>
              <TableCell>{T('k8s.namespace')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {config.namespace}
                </Typography>
              </TableCell>
            </TableRow>

            {(config as ExperimentDetail).created && (
              <TableRow>
                <TableCell>{T('table.created')}</TableCell>
                <TableCell>
                  <Typography variant="body2" color="textSecondary">
                    {format((config as ExperimentDetail).created, lang)}
                  </Typography>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </Grid>

      <Grid item md={4}>
        <Typography variant="subtitle2" gutterBottom>
          {T('newE.steps.run')}
        </Typography>

        <Table size="small">
          <TableBody>
            <TableRow>
              <TableCell>{T('newE.run.duration')}</TableCell>
              <TableCell>
                <Typography variant="body2" color="textSecondary">
                  {config.kube_object.spec.duration ? config.kube_object.spec.duration : T('newE.run.continuous')}
                </Typography>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </Grid>
    </Grid>
  )
}

export default ObjectConfiguration
