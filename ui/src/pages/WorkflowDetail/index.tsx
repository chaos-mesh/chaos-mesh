import { Box, Button, Grid } from '@material-ui/core'

import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import ReplayIcon from '@material-ui/icons/Replay'
import T from 'components/T'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => ({
  container: {
    height: 450,
  },
}))

const WorkflowDetail = () => {
  const classes = useStyles()

  return (
    <>
      <Box mb={6}>
        <Button variant="outlined" startIcon={<ReplayIcon />} onClick={() => {}}>
          {T('workflow.rerun')}
        </Button>
      </Box>
      <Grid container>
        <Grid item xs={12} md={8}>
          <Paper className={classes.container}>
            <PaperTop title={T('workflow.topology')} />
          </Paper>
        </Grid>
      </Grid>
    </>
  )
}

export default WorkflowDetail
