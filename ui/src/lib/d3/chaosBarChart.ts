import * as d3 from 'd3'

import { Experiment } from 'api/experiments.type'
import { Theme } from 'slices/settings'
import _debounce from 'lodash.debounce'
import { kind } from 'components/NewExperiment/types'

const margin = {
  top: 15,
  right: 15,
  bottom: 30,
  left: 15,
}

export default function gen({
  root,
  chaos,
  theme,
}: {
  root: HTMLElement
  chaos: { kind: Experiment['kind']; sum: number }[]
  theme: Theme
}) {
  const sumArr = chaos.map((c) => c.sum)

  let width = root.offsetWidth
  const height = root.offsetHeight

  const svg = d3
    .select(root)
    .append('svg')
    .attr('class', theme === 'light' ? 'chaos-chart' : 'chaos-chart-dark')
    .attr('width', width)
    .attr('height', height)

  const x = d3
    .scaleBand()
    .domain(kind.map((d) => d.replace('Chaos', '')))
    .range([margin.left, width - margin.right])
    .padding(0.5)
  const xAxis = d3.axisBottom(x)
  const gXAxis = svg
    .append('g')
    .attr('class', 'axis')
    .attr('transform', `translate(0, ${height - margin.bottom})`)
    .call(xAxis)

  const yHeight = height - margin.top - margin.bottom
  const domainUpperBound = (~~(d3.max(sumArr)! / 5) + 1) * 5
  const y = d3.scaleLinear().domain([0, domainUpperBound]).range([yHeight, 0])
  const yAxis = d3.axisLeft(y).ticks(5)
  // gYAxis
  svg
    .append('g')
    .attr('class', 'axis')
    .attr('transform', `translate(${margin.left}, ${margin.top})`)
    .call(yAxis)
    .call((g) => g.select('.domain').remove())
    .call((g) =>
      g
        .append('g')
        .attr('stroke-opacity', 0.5)
        .selectAll('line')
        .data(y.ticks(5))
        .join('line')
        .attr('y1', (d) => 0.5 + y(d))
        .attr('y2', (d) => 0.5 + y(d))
        .attr('x2', width - margin.right - margin.left)
    )

  const gMain = svg.append('g')
  const rects = gMain
    .selectAll()
    .data(chaos)
    .enter()
    .append('rect')
    .attr('x', (d) => x(d.kind.replace('Chaos', ''))!)
    .attr('y', (d) => y(d.sum) + margin.top)
    .attr('width', x.bandwidth())
    .attr('height', (d) => yHeight - y(d.sum))
    .attr('fill', theme === 'light' ? '#172d72' : '#9db0eb')

  function reGen() {
    const newWidth = root.offsetWidth
    width = newWidth

    svg.attr('width', width)
    x.range([margin.left, width - margin.right])
    gXAxis.call(xAxis)
    rects.attr('x', (d) => x(d.kind.replace('Chaos', ''))!).attr('width', x.bandwidth())
  }

  d3.select(window).on('resize', _debounce(reGen, 250))
}
