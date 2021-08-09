import { Grid } from '@material-ui/core'
import NewExperiment from 'components/NewExperimentNext'

const New = () => (
  <Grid container>
    <Grid item xs={12} lg={8}>
      <NewExperiment />
    </Grid>
  </Grid>
)

export default New
