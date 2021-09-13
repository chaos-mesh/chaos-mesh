import { Experiment, Selector, TableCell } from './common'
import { Grid, Table, TableBody, TableRow, Typography } from '@material-ui/core'

import { ArchiveSingle } from 'api/archives.type'
import { ExperimentSingle } from 'api/experiments.type'
import Space from 'components-mui/Space'
import StatusLabel from 'components-mui/StatusLabel'
import T from 'components/T'
import { format } from 'lib/luxon'
import { templateTypeToFieldName } from 'api/zz_generated.frontend.chaos-mesh'
import { useStoreSelector } from 'store'

type Config = ExperimentSingle | ArchiveSingle

interface ObjectConfigurationProps {
  config: Config
  inNode?: boolean
  inSchedule?: boolean
  inArchive?: boolean
  vertical?: boolean
}

const ObjectConfiguration: React.FC<ObjectConfigurationProps> = ({
  config,
  inNode,
  inSchedule,
  inArchive,
  vertical,
}) => {
  const { lang } = useStoreSelector((state) => state.settings)

  const spec = inNode ? config : config.kube_object.spec
  const experiment =
    inSchedule || inNode ? spec[templateTypeToFieldName(inSchedule ? spec.type : (config as any).templateType)] : spec

  return (
    <>
      {!inNode && (
        <Space direction="row" mb={3}>
          <Typography>{config.name}</Typography>

          {!inArchive && <StatusLabel status={(config as any).status} />}
        </Space>
      )}

      <Grid container spacing={vertical ? 3 : 0}>
        {!inNode && (
          <Grid item xs={vertical ? 12 : 3}>
            <Typography variant="subtitle2" gutterBottom>
              {T('newE.steps.basic')}
            </Typography>

            <Table size="small">
              <TableBody>
                <TableRow>
                  <TableCell>{T('k8s.namespace')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {config.namespace}
                    </Typography>
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>{T('common.uuid')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {config.uid}
                    </Typography>
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableCell>{T('table.created')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {format(config.created_at, lang)}
                    </Typography>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </Grid>
        )}

        <Grid item xs={vertical ? 12 : 3}>
          <Typography variant="subtitle2" gutterBottom>
            {T('newE.steps.scope')}
          </Typography>

          <Selector data={experiment.selector} />
        </Grid>

        <Grid item xs={vertical ? 12 : 3}>
          <Typography variant="subtitle2" gutterBottom>
            {T('experiments.single')}
          </Typography>

          <Experiment
            kind={inNode ? (config as any).templateType : inSchedule ? spec.type : config.kind}
            data={experiment}
          />
        </Grid>

        <Grid item xs={vertical ? 12 : 3}>
          <Typography variant="subtitle2" gutterBottom>
            {T('newE.steps.run')}
          </Typography>

          <Table size="small">
            <TableBody>
              {!inSchedule && (
                <TableRow>
                  <TableCell>{T(inNode ? 'newW.node.deadline' : 'common.duration')}</TableCell>
                  <TableCell>
                    <Typography variant="body2" color="textSecondary">
                      {inNode ? (config as any).deadline : spec.duration ? spec.duration : T('newE.run.continuous')}
                    </Typography>
                  </TableCell>
                </TableRow>
              )}
              {inSchedule && (
                <>
                  <TableRow>
                    <TableCell>{T('schedules.single')}</TableCell>
                    <TableCell>
                      <Typography variant="body2" color="textSecondary">
                        {spec.schedule}
                      </Typography>
                    </TableCell>
                  </TableRow>
                  {spec.historyLimit && (
                    <TableRow>
                      <TableCell>{T('newS.basic.historyLimit')}</TableCell>
                      <TableCell>
                        <Typography variant="body2" color="textSecondary">
                          {spec.historyLimit}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                  {spec.concurrencyPolicy && (
                    <TableRow>
                      <TableCell>{T('newS.basic.concurrencyPolicy')}</TableCell>
                      <TableCell>
                        <Typography variant="body2" color="textSecondary">
                          {spec.concurrencyPolicy}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                  {spec.startingDeadlineSeconds && (
                    <TableRow>
                      <TableCell>{T('newS.basic.startingDeadlineSeconds')}</TableCell>
                      <TableCell>
                        <Typography variant="body2" color="textSecondary">
                          {spec.startingDeadlineSeconds}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  )}
                </>
              )}
            </TableBody>
          </Table>
        </Grid>
      </Grid>
    </>
  )
}

export default ObjectConfiguration
