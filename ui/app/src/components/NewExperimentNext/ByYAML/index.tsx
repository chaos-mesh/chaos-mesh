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
import Paper from '@/mui-extends/Paper'
import Space from '@/mui-extends/Space'
import { useStoreDispatch } from '@/store'
import PublishIcon from '@mui/icons-material/Publish'
import { Button, Typography } from '@mui/material'
import { type Editor } from 'ace-builds'
import yaml from 'js-yaml'
import { lazy, useState } from 'react'
import { useIntl } from 'react-intl'

import { setAlert } from '@/slices/globalStatus'

import i18n from '@/components/T'
import YAML from '@/components/YAML'

const YAMLEditor = lazy(() => import('@/components/YAMLEditor'))

interface ByYAMLProps {
  callback?: (data: any) => void
}

const ByYAML: ReactFCWithChildren<ByYAMLProps> = ({ callback }) => {
  const intl = useIntl()

  const dispatch = useStoreDispatch()

  const [empty, setEmpty] = useState(true)
  const [yamlEditor, setYAMLEditor] = useState<Editor>()

  const onChange = (value: string) => setEmpty(value === '')

  const handleUploadYAMLCallback = (y: any) => yamlEditor?.setValue(y)

  const handleSubmit = () => {
    const data = yaml.load(yamlEditor!.getValue())

    callback && callback(data)

    dispatch(
      setAlert({
        type: 'success',
        message: i18n('confirm.success.load', intl),
      }),
    )
  }

  return (
    <Space spacing={6}>
      <Typography variant="body2" color="textSecondary">
        {i18n('newE.byYAMLDesc')}
      </Typography>
      <Paper sx={{ height: 600, p: 0 }}>
        <YAMLEditor mountEditor={setYAMLEditor} aceProps={{ onChange }} />
      </Paper>
      <Space direction="row" justifyContent="flex-end">
        <YAML callback={handleUploadYAMLCallback} />
        <Button
          variant="contained"
          color="primary"
          startIcon={<PublishIcon />}
          size="small"
          disabled={empty}
          onClick={handleSubmit}
        >
          {i18n('common.submit')}
        </Button>
      </Space>
    </Space>
  )
}

export default ByYAML
