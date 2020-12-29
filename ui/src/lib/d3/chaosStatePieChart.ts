import * as d3 from 'd3'

import { IntlShape } from 'react-intl'
import { StateOfExperiments } from 'api/experiments.type'
import { Theme } from 'slices/settings'

interface Data {
  name: string
  value: number
}

export default function gen({
  root,
  chaosStatus,
  intl,
  theme,
}: {
  root: HTMLElement
  chaosStatus: StateOfExperiments
  intl: IntlShape
  theme: Theme
}) {
  let width = root.offsetWidth
  const height = root.offsetHeight

  const svg = d3
    .select(root)
    .append('svg')
    .attr('class', theme === 'light' ? 'chaos-chart' : 'chaos-chart-dark')
    .attr('width', width)
    .attr('height', height + 12)
    .attr('style', 'position: relative; top: -12px;')

  svg
    .append('g')
    .attr('class', 'slices')
    .attr('transform', `translate(${width / 2}, ${height / 2})`)
  svg
    .append('g')
    .attr('class', 'labels')
    .attr('transform', `translate(${width / 2}, ${height / 2})`)

  const key = (d: d3.PieArcDatum<Data>) => d.data.name
  const pie = d3
    .pie<Data>()
    .sort(null)
    .value((d) => d.value)
  const radius = Math.min(width, height) / 2
  const arc = d3
    .arc<d3.PieArcDatum<Data>>()
    .innerRadius(radius * 0.6)
    .outerRadius(radius * 0.75)
  const arcLabel = d3.arc<d3.PieArcDatum<Data>>().innerRadius(radius).outerRadius(radius)

  update(chaosStatus)

  function update(data: typeof chaosStatus) {
    const processed = Object.entries(data).map((d) => ({ name: d[0], value: d[1] }))

    const color = d3
      .scaleOrdinal<string>()
      .domain(processed.map((d) => d.name))
      .range(d3.schemeTableau10)

    const slice = svg
      .select('.slices')
      .selectAll<SVGPathElement, d3.PieArcDatum<Data>>('path.slice')
      .data<d3.PieArcDatum<Data>>(pie(processed), key)

    slice
      .join(
        (enter) => enter.append('path'),
        (update) =>
          update.call((update) =>
            update
              .transition()
              .duration(1500)
              .attrTween('d', function (this: any, d) {
                this._current = this._current || d
                const interpolate = d3.interpolate(this._current, d)
                this._current = interpolate(0)

                return (t) => arc(interpolate(t)) as string
              })
          ),
        (exit) => exit.remove()
      )
      .attr('class', 'slice')
      .style('fill', (d) => color(d.data.name))

    const text = svg
      .select('.labels')
      .selectAll<SVGTextElement, d3.PieArcDatum<Data>>('text')
      .data<d3.PieArcDatum<Data>>(pie(processed), key)

    text
      .join(
        (enter) => enter.append('text'),
        (update) =>
          update.call((update) =>
            update
              .transition()
              .duration(1500)
              .attrTween('transform', function (this: any, d) {
                this._current = this._current || d
                const interpolate = d3.interpolate(this._current, d)
                this._current = interpolate(0)

                return (t) => {
                  const d2 = interpolate(t)
                  const pos = arcLabel.centroid(d2)

                  return `translate(${pos})`
                }
              })
          ),
        (exit) => exit.remove()
      )
      .style('text-anchor', 'middle')
      .text((d) =>
        d.value > 0
          ? `${d.data.value} ${intl.formatMessage({ id: `experiments.state.${d.data.name.toLowerCase()}` })}`
          : ''
      )
  }

  return update
}
