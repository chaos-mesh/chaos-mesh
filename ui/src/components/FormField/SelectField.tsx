import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { Field, getIn, useFormikContext } from 'formik'

import { Experiment } from 'components/NewExperiment/types'
import React from 'react'
import clsx from 'clsx'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles({
  muiSelectRoot: {
    '& .MuiSelect-root': {
      padding: 6,
      paddingTop: 8,
    },
  },
})

const SelectField: React.FC<TextFieldProps & { multiple?: boolean }> = ({ multiple = false, ...props }) => {
  const classes = useStyles()

  const { values, setFieldValue } = useFormikContext<Experiment>()

  const onDelete = (val: string) => () =>
    setFieldValue(
      props.name!,
      getIn(values, props.name!).filter((d: string) => d !== val)
    )

  const SelectProps = {
    multiple,
    renderValue: multiple
      ? (selected: any) => (
          <Box display="flex" flexWrap="wrap">
            {(selected as string[]).map((val) => (
              <Box key={val} m={0.5}>
                <Chip
                  style={{ height: 24, margin: 1 }}
                  label={val}
                  color="primary"
                  onDelete={onDelete(val)}
                  onMouseDown={(e) => e.stopPropagation()}
                />
              </Box>
            ))}
          </Box>
        )
      : undefined,
  }

  return (
    <Box mb={3}>
      <Field
        {...props}
        as={TextField}
        className={clsx(multiple && classes.muiSelectRoot, props.className)}
        variant="outlined"
        select
        margin="dense"
        fullWidth
        SelectProps={SelectProps}
      />
    </Box>
  )
}

export default SelectField
