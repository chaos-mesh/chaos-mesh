import { Box, Typography } from '@material-ui/core'
import { useStoreDispatch, useStoreSelector } from 'store'

import DoneAllIcon from '@material-ui/icons/DoneAll'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import React from 'react'
import { Submit } from 'components/FormField'
import T from 'components/T'
import api from 'api'
import { parseSubmit } from 'lib/formikhelpers'
import { resetNewExperiment } from 'slices/experiments'
import { setAlert } from 'slices/globalStatus'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'

interface Step3Props {
  onSubmit?: (experiment: { target: any; basic: any }) => void
}

const Step3: React.FC<Step3Props> = ({ onSubmit }) => {
  const state = useStoreSelector((state) => state)
  const { step1, step2, basic, target } = state.experiments
  const { debugMode } = state.settings
  const dispatch = useStoreDispatch()

  const history = useHistory()
  const intl = useIntl()

  const submitExperiment = () => {
    const parsedValues = parseSubmit({
      ...basic,
      target,
    })

    if (process.env.NODE_ENV === 'development' || debugMode) {
      console.debug('Debug parsedValues:', parsedValues)
    }

    if (!debugMode) {
      if (onSubmit) {
        onSubmit({ target, basic })
      } else {
        api.experiments
          .newExperiment(parsedValues)
          .then(() => {
            dispatch(
              setAlert({
                type: 'success',
                message: intl.formatMessage({ id: 'confirm.createSuccessfully' }),
              })
            )

            dispatch(resetNewExperiment())

            history.push('/experiments')
          })
          .catch(console.error)
      }
    }
  }

  return (
    <>
      {step1 && step2 && (
        <Paper>
          <PaperTop title={T('common.submit')} />
          <Box textAlign="center">
            <DoneAllIcon fontSize="large" />
            <Typography>{T('newE.complete')}</Typography>
            <Submit onClick={submitExperiment} />
          </Box>
        </Paper>
      )}
    </>
  )
}

export default Step3
