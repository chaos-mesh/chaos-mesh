import { LabelAccessorFunction, PieTooltipProps, ResponsivePie } from '@nivo/pie'
import React, { useEffect, useState } from 'react'

import { Box } from '@material-ui/core'
import NotFound from 'components-mui/NotFound'
import { StateOfExperiments } from 'api/experiments.type'
import T from 'components/T'
import api from 'api'
import { schemeTableau10 } from 'd3-scale-chromatic'
import { useIntl } from 'react-intl'
import { useTheme } from '@material-ui/core/styles'

interface SingleData {
  id: keyof StateOfExperiments
  value: number
}

interface TotalStateProps {
  className?: string
}

const TotalState: React.FC<TotalStateProps> = (props) => {
  const intl = useIntl()
  const theme = useTheme()

  const [s, setS] = useState<SingleData[]>([])

  const fetchState = () => {
    api.experiments
      .state()
      .then((resp) => setS(Object.entries(resp.data).map(([k, v]) => ({ id: k as any, value: v === 0 ? 0.01 : v }))))
      .catch(console.error)
  }

  const radialLabel: LabelAccessorFunction<SingleData> = (d) =>
    d.value + ' ' + intl.formatMessage({ id: `experiments.state.${d.id.toString().toLowerCase()}` })

  const tooltip = ({ datum }: PieTooltipProps<SingleData>) => (
    <Box
      display="flex"
      alignItems="center"
      p={1.5}
      style={{ background: theme.palette.background.default, fontSize: theme.typography.caption.fontSize }}
    >
      <Box mr={1.5} style={{ width: 12, height: 12, background: datum.color, borderRadius: 50 }} />
      {(datum.value < 1 ? 0 : datum.value) +
        ' ' +
        intl.formatMessage({ id: `experiments.state.${datum.id.toString().toLowerCase()}` })}
    </Box>
  )

  useEffect(() => {
    fetchState()

    const id = setInterval(fetchState, 15000)

    return () => clearInterval(id)
  }, [])

  return (
    <div className={props.className}>
      {s.some((d) => d.value >= 1) ? (
        <ResponsivePie
          data={s}
          margin={{ top: 15, right: 15, bottom: 15, left: 15 }}
          colors={schemeTableau10 as any}
          innerRadius={0.75}
          padAngle={0.25}
          cornerRadius={4}
          radialLabel={radialLabel}
          radialLabelsSkipAngle={4}
          radialLabelsLinkDiagonalLength={8}
          radialLabelsLinkHorizontalLength={12}
          radialLabelsLinkColor={{
            from: 'color',
          }}
          radialLabelsTextColor={theme.palette.text.primary}
          enableSliceLabels={false}
          tooltip={tooltip}
        />
      ) : (
        <NotFound>{T('experiments.noExperimentsFound')}</NotFound>
      )}
    </div>
  )
}

export default TotalState
