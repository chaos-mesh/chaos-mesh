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
import { Box, BoxProps, Typography } from '@material-ui/core'

interface PaperTopProps {
  title: string | JSX.Element
  subtitle?: string | JSX.Element
  boxProps?: BoxProps
}

const PaperTop: React.FC<PaperTopProps> = ({ title, subtitle, boxProps, children }) => (
  <Box {...boxProps} display="flex" justifyContent="space-between" width="100%">
    <div>
      <Typography component="div" gutterBottom={subtitle ? true : false}>
        {title}
      </Typography>
      {subtitle && (
        <Typography variant="body2" color="textSecondary">
          {subtitle}
        </Typography>
      )}
    </div>
    {children}
  </Box>
)

export default PaperTop
