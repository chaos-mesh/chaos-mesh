import { resetNewExperiment, setScheduleSpecific } from 'slices/experiments'
import { useStoreDispatch, useStoreSelector } from 'store'

import { Grid } from '@material-ui/core'
import NewExperiment from 'components/NewExperimentNext'
import api from 'api'
import { setAlert } from 'slices/globalStatus'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'

const New = () => {
  const history = useHistory()
  const intl = useIntl()

  const { scheduleSpecific } = useStoreSelector((state) => state.experiments)
  const dispatch = useStoreDispatch()

  const onSubmit = ({ target, basic }: any) => {
    const data = {
      ...basic,
      target,
      ...scheduleSpecific,
    }

    api.schedules
      .newSchedule(data)
      .then(() => {
        dispatch(
          setAlert({
            type: 'success',
            message: intl.formatMessage({ id: 'confirm.createSuccessfully' }),
          })
        )

        dispatch(resetNewExperiment())
        dispatch(setScheduleSpecific({} as any))

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
