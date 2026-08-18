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
import Paper from '@/mui-extends/Paper'
import PaperTop from '@/mui-extends/PaperTop'
import {
  Box,
  Chip,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
} from '@mui/material'

import { T } from '@/components/T'

interface RecordEvent {
  message?: string
  operation?: string
  timestamp?: string
  type?: string
}

interface ContainerRecord {
  events?: RecordEvent[]
  id?: string
  injectedCount?: number
  phase?: string
  recoveredCount?: number
  selectorKey?: string
}

export interface ExperimentChaosStatus {
  experiment?: {
    containerRecords?: ContainerRecord[]
  }
}

interface ExperimentRecordsProps {
  status?: ExperimentChaosStatus
}

const ExperimentRecords: ReactFCWithChildren<ExperimentRecordsProps> = ({ status }) => {
  const records = status?.experiment?.containerRecords

  if (!records?.length) {
    return null
  }

  return (
    <Paper>
      <PaperTop title={<T id="experiments.records.title" />} />
      <TableContainer sx={{ mt: 2, maxHeight: 480 }}>
        <Table stickyHeader size="small">
          <TableHead>
            <TableRow>
              <TableCell>
                <T id="experiments.records.target" />
              </TableCell>
              <TableCell>
                <T id="experiments.records.selector" />
              </TableCell>
              <TableCell>
                <T id="experiments.records.phase" />
              </TableCell>
              <TableCell align="right">
                <T id="experiments.records.injected" />
              </TableCell>
              <TableCell align="right">
                <T id="experiments.records.recovered" />
              </TableCell>
              <TableCell>
                <T id="experiments.records.events" />
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {records.map((record, recordIndex) => (
              <TableRow key={`${record.selectorKey}-${record.id}-${recordIndex}`}>
                <TableCell>
                  <Typography variant="body2" sx={{ fontFamily: 'monospace', overflowWrap: 'anywhere' }}>
                    {record.id}
                  </Typography>
                </TableCell>
                <TableCell>{record.selectorKey}</TableCell>
                <TableCell>
                  <Chip
                    color={record.phase === 'Injected' ? 'success' : 'default'}
                    label={record.phase}
                    size="small"
                    variant="outlined"
                  />
                </TableCell>
                <TableCell align="right">{record.injectedCount ?? 0}</TableCell>
                <TableCell align="right">{record.recoveredCount ?? 0}</TableCell>
                <TableCell>
                  {record.events?.length ? (
                    <Box component="details">
                      <Box component="summary" sx={{ cursor: 'pointer' }}>
                        <T id="experiments.records.eventCount" values={{ count: record.events.length }} />
                      </Box>
                      <Stack spacing={1} sx={{ mt: 1 }}>
                        {record.events.map((event, eventIndex) => (
                          <Box key={`${event.timestamp}-${event.operation}-${eventIndex}`}>
                            <Stack direction="row" spacing={1} alignItems="center">
                              <Chip
                                color={event.type === 'Failed' ? 'error' : 'success'}
                                label={`${event.operation || ''} ${event.type || ''}`.trim()}
                                size="small"
                              />
                              {event.timestamp && (
                                <Typography variant="caption" color="text.secondary">
                                  {event.timestamp}
                                </Typography>
                              )}
                            </Stack>
                            {event.message && (
                              <Typography variant="caption" color="text.secondary" display="block" sx={{ mt: 0.5 }}>
                                {event.message}
                              </Typography>
                            )}
                          </Box>
                        ))}
                      </Stack>
                    </Box>
                  ) : (
                    <Typography variant="body2" color="text.secondary">
                      <T id="experiments.records.noEvents" />
                    </Typography>
                  )}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Paper>
  )
}

export default ExperimentRecords
