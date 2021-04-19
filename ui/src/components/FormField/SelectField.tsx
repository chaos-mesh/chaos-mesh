import { Box, Chip, TextField, TextFieldProps } from '@material-ui/core'
import { Field, getIn, useFormikContext } from 'formik'

import { Experiment } from 'components/NewExperiment/types'
import React from 'react'
import clsx from 'clsx'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles({
  root: {
    '& .MuiSelect-root': {
      padding: 6,
      paddingTop: 8,
    },
  },
})

const SelectField: React.FC<TextFieldProps & { multiple?: boolean; mb?: number }> = ({
  multiple = false,
  mb = 1.5,
  ...props
}) => {
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

  const rendered = (
    <Field
      {...props}
      className={clsx(multiple && classes.root, props.className)}
      as={TextField}
      variant="outlined"
      select
      margin="dense"
      fullWidth
      SelectProps={SelectProps}
    />
  )

  return mb > 0 ? <Box mb={mb}>{rendered}</Box> : rendered
}

export default SelectField
