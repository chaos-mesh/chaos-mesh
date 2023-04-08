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
import { Box, BoxProps } from '@mui/material'
import { useTheme } from '@mui/material/styles'
import { PropertyAccessor } from '@nivo/core'
import { ComputedDatum, PieTooltipProps, ResponsivePie } from '@nivo/pie'
import { useGetExperimentsState } from 'openapi'
import { StatusAllChaosStatus } from 'openapi/index.schemas'
import { useState } from 'react'
import { useIntl } from 'react-intl'

import NotFound from 'components/NotFound'
import i18n from 'components/T'

interface SingleData {
  id: keyof StatusAllChaosStatus
  label: string
  value: number
}

const TotalStatus: React.FC<BoxProps> = (props) => {
  const intl = useIntl()
  const theme = useTheme()

  const [state, setState] = useState<SingleData[]>([])

  const arcLinkLabel: PropertyAccessor<ComputedDatum<SingleData>, string> = (d) =>
    d.value + ' ' + i18n(`status.${d.id}`, intl)

  const tooltip = ({ datum }: PieTooltipProps<SingleData>) => (
    <Box
      display="flex"
      alignItems="center"
      p={1.5}
      style={{ background: theme.palette.background.default, fontSize: theme.typography.caption.fontSize }}
    >
      <Box mr={1.5} style={{ width: 12, height: 12, background: datum.color }} />
      {(datum.value < 1 ? 0 : datum.value) + ' ' + i18n(`status.${datum.id}`, intl)}
    </Box>
  )

  useGetExperimentsState(undefined, {
    query: {
      onSuccess(data) {
        setState(
          (Object.entries(data) as [keyof StatusAllChaosStatus, number][]).map(([k, v]) => ({
            id: k,
            label: i18n(`status.${k}`, intl),
            value: v === 0 ? 0.01 : v,
          }))
        )
      },
    },
  })

  return (
    <Box {...props}>
      {state.some((d) => d.value >= 1) ? (
        <ResponsivePie
          data={state}
          margin={{ top: 15, bottom: 60 }}
          innerRadius={0.75}
          padAngle={0.25}
          cornerRadius={4}
          enableArcLabels={false}
          arcLinkLabel={arcLinkLabel}
          arcLinkLabelsSkipAngle={4}
          arcLinkLabelsDiagonalLength={8}
          arcLinkLabelsStraightLength={12}
          arcLinkLabelsColor={{
            from: 'color',
          }}
          arcLinkLabelsTextColor={theme.palette.text.primary}
          tooltip={tooltip}
          activeInnerRadiusOffset={2}
          activeOuterRadiusOffset={2}
          legends={[
            {
              anchor: 'bottom',
              direction: 'row',
              itemWidth: 75,
              itemHeight: 30,
              translateY: 60,
            },
          ]}
        />
      ) : (
        <NotFound>{i18n('experiments.notFound')}</NotFound>
      )}
    </Box>
  )
}

export default TotalStatus
