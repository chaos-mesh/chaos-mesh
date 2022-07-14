/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import FileOpenIcon from '@mui/icons-material/FileOpen'
import LoadingButton, { LoadingButtonProps } from '@mui/lab/LoadingButton'
import { useState } from 'react'

import { useStoreDispatch } from 'store'

import { setAlert } from 'slices/globalStatus'

import { T } from 'components/T'

interface YAMLProps {
  callback: (y: any) => void
  ButtonProps?: LoadingButtonProps<'label'>
}

const YAML: React.FC<YAMLProps> = ({ children, callback, ButtonProps }) => {
  const [loading, setLoading] = useState(false)

  const dispatch = useStoreDispatch()

  const handleUploadYAML = (e: React.ChangeEvent<HTMLInputElement>) => {
    setLoading(true)

    const f = e.target.files![0]

    const reader = new FileReader()
    reader.onload = function () {
      const y = reader.result

      callback(y)

      dispatch(
        setAlert({
          type: 'success',
          message: <T id="confirm.success.load" />,
        })
      )

      setLoading(false)
    }
    reader.readAsText(f)
  }

  return (
    <LoadingButton
      {...ButtonProps}
      component="label"
      loading={loading}
      variant="outlined"
      size="small"
      startIcon={<FileOpenIcon />}
    >
      {children || <T id="common.upload" />}
      <input type="file" hidden onChange={handleUploadYAML} />
    </LoadingButton>
  )
}

export default YAML
