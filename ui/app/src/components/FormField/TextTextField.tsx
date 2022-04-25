import { Box, Button, FormHelperText, IconButton, Typography } from '@mui/material'
import { getIn, useFormikContext } from 'formik'

import AddCircleTwoToneIcon from '@mui/icons-material/AddCircleTwoTone'
import LabelField from './LabelField'
import RemoveCircleTwoToneIcon from '@mui/icons-material/RemoveCircleTwoTone'
import Space from '@ui/mui-extends/esm/Space'
import TextField from './TextField'
import _ from 'lodash'

interface TextTextFieldProps {
  name: string
  label: React.ReactNode
  helperText?: React.ReactNode
  valueLabeled?: boolean
}

export default function TextTextField({ name, label, helperText, valueLabeled }: TextTextFieldProps) {
  const { values, setFieldValue } = useFormikContext()
  const fieldValue = getIn(values, name)
  const entries = Object.entries(fieldValue)

  const handleAddKV = (n: number) => () => {
    setFieldValue(name, {
      ...fieldValue,
      ['key' + n]: { key: '', value: valueLabeled ? [] : '' },
    })
  }

  const handleRemoveKV = (key: string) => () => {
    setFieldValue(name, _.omit(fieldValue, key))
  }

  return (
    <>
      <Box>
        <Typography variant="body2" fontWeight={500}>
          {label}
        </Typography>
        {helperText && <FormHelperText>{helperText}</FormHelperText>}
      </Box>
      {entries.length > 0 ? (
        entries.map(([key], i) => (
          <Space key={key} direction="row">
            <TextField fast name={`${name}.${key}.key`} />
            {valueLabeled ? (
              <LabelField name={`${name}.${key}.value`} sx={{ width: 300 }} />
            ) : (
              <TextField fast name={`${name}.${key}.value`} />
            )}
            <IconButton color="error" onClick={handleRemoveKV(key)}>
              <RemoveCircleTwoToneIcon />
            </IconButton>
            {i === entries.length - 1 && (
              <IconButton color="primary" onClick={handleAddKV(i + 1)}>
                <AddCircleTwoToneIcon />
              </IconButton>
            )}
          </Space>
        ))
      ) : (
        <Box>
          <Button variant="contained" onClick={handleAddKV(0)}>
            {`Add Key/Value${valueLabeled ? 's' : ''} Pair`}
          </Button>
        </Box>
      )}
    </>
  )
}
