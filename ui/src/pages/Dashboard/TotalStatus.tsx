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
import { Box, BoxProps } from '@material-ui/core'
import { ComputedDatum, PieTooltipProps, ResponsivePie } from '@nivo/pie'
import { useEffect, useState } from 'react'

import NotFound from 'components-mui/NotFound'
import { PropertyAccessor } from '@nivo/core'
import { StatusOfExperiments } from 'api/experiments.type'
import T from 'components/T'
import api from 'api'
import { schemeTableau10 } from 'd3-scale-chromatic'
import { useIntl } from 'react-intl'
import { useTheme } from '@material-ui/core/styles'

interface SingleData {
  id: keyof StatusOfExperiments
  label: string
  value: number
}

const TotalStatus: React.FC<BoxProps> = (props) => {
  const intl = useIntl()
  const theme = useTheme()

  const [s, setS] = useState<SingleData[]>([])

  const arcLinkLabel: PropertyAccessor<ComputedDatum<SingleData>, string> = (d) =>
    d.value + ' ' + T(`status.${d.id}`, intl)

  const tooltip = ({ datum }: PieTooltipProps<SingleData>) => (
    <Box
      display="flex"
      alignItems="center"
      p={1.5}
      style={{ background: theme.palette.background.default, fontSize: theme.typography.caption.fontSize }}
    >
      <Box mr={1.5} style={{ width: 12, height: 12, background: datum.color }} />
      {(datum.value < 1 ? 0 : datum.value) + ' ' + T(`status.${datum.id}`, intl)}
    </Box>
  )

  useEffect(() => {
    const fetchState = () => {
      api.experiments
        .state()
        .then((resp) =>
          setS(
            (Object.entries(resp.data) as [keyof StatusOfExperiments, number][]).map(([k, v]) => ({
              id: k,
              label: T(`status.${k}`, intl),
              value: v === 0 ? 0.01 : v,
            }))
          )
        )
        .catch(console.error)
    }

    fetchState()

    const id = setInterval(fetchState, 12000)

    return () => clearInterval(id)
  }, [intl])

  return (
    <Box {...props}>
      {s.some((d) => d.value >= 1) ? (
        <ResponsivePie
          data={s}
          margin={{ top: 15, bottom: 60 }}
          colors={schemeTableau10 as any}
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
        <NotFound>{T('experiments.notFound')}</NotFound>
      )}
    </Box>
  )
}

export default TotalStatus
