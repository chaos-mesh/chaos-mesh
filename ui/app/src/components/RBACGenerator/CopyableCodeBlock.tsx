/*
 * Copyright 2026 Chaos Mesh Authors.
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
import { Box, Button } from '@mui/material'
import { styled } from '@mui/material/styles'

interface CopyableCodeBlockProps {
  text: string
  copyLabel: string
  onCopy: (text: string) => void
  height?: number
  singleLine?: boolean
}

const CodeBlock = styled('pre')(({ theme }) => ({
  padding: theme.spacing(3),
  background: theme.palette.background.default,
  borderRadius: 4,
}))

interface CopyButtonProps {
  singleLine: boolean
}

const CopyButton = styled(Box, {
  shouldForwardProp: (prop) => prop !== 'singleLine',
})<CopyButtonProps>(({ theme, singleLine }) => ({
  position: 'absolute',
  top: singleLine ? '50%' : theme.spacing(6),
  right: singleLine ? theme.spacing(1) : theme.spacing(3),
  transform: singleLine ? 'translateY(-50%)' : undefined,
}))

const CopyableCodeBlock = ({ text, copyLabel, onCopy, height, singleLine = false }: CopyableCodeBlockProps) => (
  <Box position="relative">
    <CodeBlock
      style={
        singleLine
          ? { overflowX: 'auto', whiteSpace: 'pre', paddingRight: 80 }
          : { height, overflow: height ? 'auto' : undefined, whiteSpace: 'pre-wrap' }
      }
    >
      {text}
    </CodeBlock>
    <CopyButton singleLine={singleLine}>
      <Button onClick={() => onCopy(text)}>{copyLabel}</Button>
    </CopyButton>
  </Box>
)

export default CopyableCodeBlock
