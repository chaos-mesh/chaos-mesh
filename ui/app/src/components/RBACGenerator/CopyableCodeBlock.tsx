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

const CodePanel = styled(Box)(({ theme }) => ({
  display: 'flex',
  gap: theme.spacing(1),
  background: theme.palette.background.default,
  borderRadius: 4,
}))

const CodeBlock = styled('pre')(({ theme }) => ({
  flex: '1 1 0',
  minInlineSize: 0,
  margin: 0,
  padding: theme.spacing(3),
}))

const CopyableCodeBlock = ({ text, copyLabel, onCopy, height, singleLine = false }: CopyableCodeBlockProps) => (
  <CodePanel alignItems={singleLine ? 'center' : 'flex-start'}>
    <CodeBlock
      style={
        singleLine
          ? { overflowX: 'auto', whiteSpace: 'pre' }
          : { height, overflow: height ? 'auto' : undefined, whiteSpace: 'pre-wrap' }
      }
    >
      {text}
    </CodeBlock>
    <Button sx={{ flex: '0 0 auto', mr: singleLine ? 1 : 3, mt: singleLine ? 0 : 3 }} onClick={() => onCopy(text)}>
      {copyLabel}
    </Button>
  </CodePanel>
)

export default CopyableCodeBlock
