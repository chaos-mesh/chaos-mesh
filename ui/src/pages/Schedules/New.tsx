import { Grid } from '@material-ui/core'
import NewExperiment from 'components/NewExperimentNext'
import T from 'components/T'
import api from 'api'
import { resetNewExperiment } from 'slices/experiments'
import { setAlert } from 'slices/globalStatus'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'
import { useStoreDispatch } from 'store'

const New = () => {
  const history = useHistory()
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const onSubmit = (parsedValues: any) => {
    api.schedules
      .newSchedule(parsedValues)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: T('confirm.success.create', intl),
          })
        )

        dispatch(resetNewExperiment())

        history.push('/schedules')
      })
      .catch(console.error)
  }

  return (
    <Grid container>
      <Grid item xs={12} lg={8}>
        <NewExperiment inSchedule={true} onSubmit={onSubmit} />
      </Grid>
    </Grid>
  )
}

export default New
