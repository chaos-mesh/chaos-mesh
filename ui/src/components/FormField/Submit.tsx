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
import { Box, Button, ButtonProps } from '@material-ui/core'

import PublishIcon from '@material-ui/icons/Publish'
import T from 'components/T'

export default function Submit({ mt = 6, onClick, ...rest }: ButtonProps & { mt?: number }) {
  return (
    <Box mt={mt} textAlign="right">
      <Button
        type={onClick ? undefined : 'submit'}
        variant="contained"
        startIcon={<PublishIcon />}
        onClick={onClick}
        {...rest}
      >
        {T('common.submit')}
      </Button>
    </Box>
  )
}
