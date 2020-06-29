import * as d3 from 'd3'

import { Event } from 'api/events.type'
import day from 'lib/dayjs'
import insertCommonStyle from './insertCommonStyle'
import wrapText from './wrapText'

const margin = {
  top: 0,
  right: 15,
  bottom: 30,
  left: 15,
}

export default function gen({
  root,
  events,
  selectEvent,
}: {
  root: HTMLElement
  events: Event[]
  selectEvent?: (e: Event) => void
}) {
  insertCommonStyle()

  let width = root.offsetWidth
  const height = root.offsetHeight

  const svg = d3
    .select(root)
    .append('svg')
    .attr('class', 'chaos-events-chart')
    .attr('width', width)
    .attr('height', height)

  const now = day(events[events.length - 1].StartTime).add(0.5, 'h')

  const x = d3
    .scaleLinear()
    .domain([now.subtract(1, 'h'), now])
    .range([margin.left, width - margin.right])
  const xAxis = d3
    .axisBottom(x)
    .ticks(6)
    .tickFormat(d3.timeFormat('%m-%d %H:%M') as (dv: Date | { valueOf(): number }, i: number) => string)
  const gXAxis = svg
    .append('g')
    .attr('class', 'axis')
    .attr('transform', `translate(0, ${height - margin.bottom})`)
    .call(xAxis)

  // Wrap long text, also used in zoom() and reGen()
  svg.selectAll('.tick text').call(wrapText, 30)

  const allExperiments = [...new Set(events.map((d) => d.Experiment))]
  const y = d3
    .scaleBand()
    .domain(allExperiments)
    .range([0, height - margin.top - margin.bottom])
    .padding(0.25)
  const yAxis = d3.axisLeft(y).tickFormat('' as any)
  // gYAxis
  svg.append('g').attr('class', 'axis').attr('transform', `translate(${margin.left}, ${margin.top})`).call(yAxis)

  // clipX
  svg
    .append('clipPath')
    .attr('id', 'clip-x-axis')
    .append('rect')
    .attr('x', margin.left)
    .attr('y', 0)
    .attr('width', width - margin.left - margin.right)
    .attr('height', height - margin.bottom)
  const gMain = svg.append('g').attr('clip-path', 'url(#clip-x-axis)')

  const colorPalette = d3
    .scaleOrdinal<string, string>()
    .domain(events.map((d) => d.Experiment))
    .range(d3.schemeTableau10)

  const legendsRoot = d3.select(document.createElement('div')).attr('class', 'chaos-events-legends')
  const legends = legendsRoot.selectAll().data(allExperiments).enter().append('div')
  legends
    .insert('div')
    .attr('style', 'color: rgba(0, 0, 0, 0.72); font-size: 0.625rem;')
    .text((d) => d)
  legends
    .append('div')
    .attr(
      'style',
      (d) => `width: 12px; height: 12px; margin-left: 8px; background: ${colorPalette(d)}; border-radius: 3px;`
    )

  function genRectWidth(d: Event) {
    let width = d.FinishTime ? x(day(d.FinishTime)) - x(day(d.StartTime)) : x(day()) - x(day(d.StartTime))

    if (width === 0) {
      width = 20
    }

    return width
  }

  const rects = gMain
    .selectAll()
    .data(events)
    .enter()
    .append('rect')
    .attr('x', (d) => x(day(d.StartTime)))
    .attr('y', (d) => y(d.Experiment)! + margin.top)
    .attr('width', genRectWidth)
    .attr('height', y.bandwidth())
    .attr('fill', (d) => colorPalette(d.Experiment))
    .style('cursor', 'pointer')

  const zoom = d3.zoom().scaleExtent([0.5, 5]).on('zoom', zoomd)
  svg.call(zoom as any)
  function zoomd() {
    const eventTransform = d3.event.transform

    const newX = eventTransform.rescaleX(x)

    gXAxis.call(xAxis.scale(newX))
    svg.selectAll('.tick text').call(wrapText, 30)
    rects.attr('x', (d) => newX(day(d.StartTime))).attr('width', genRectWidth)
  }

  const tooltip = d3
    .select(document.createElement('div'))
    .attr('class', 'chaos-event-tooltip')
    .call(createTooltip as any)

  function createTooltip(el: d3.Selection<HTMLElement, any, any, any>) {
    el.style('position', 'absolute')
      .style('top', 0)
      .style('left', 0)
      .style('padding', '0.25rem 0.75rem')
      .style('background', '#fff')
      .style('font', '1rem')
      .style('border', '1px solid rgba(0, 0, 0, 0.12)')
      .style('border-radius', '4px')
      .style('opacity', 0)
      .style('transition', 'top 0.25s ease, left 0.25s ease')
      .style('z-index', 999)
  }

  function genTooltipContent(d: Event) {
    return `<b>Experiment: ${d.Experiment}</b>
            <br />
            <b>Status: ${d.FinishTime ? 'Finished' : 'Running'}</b>
            <br />
            <br />
            <span style="color: rgba(0, 0, 0, 0.67);">Start Time: ${day(d.StartTime).format(
              'YYYY-MM-DD HH:mm:ss A'
            )}</span>
            <br />
            ${
              d.FinishTime
                ? `<span style="color: rgba(0, 0, 0, 0.67);">Finish Time: ${day(d.FinishTime).format(
                    'YYYY-MM-DD HH:mm:ss A'
                  )}</span>`
                : ''
            }
            `
  }

  rects
    .on('click', function (d) {
      if (typeof selectEvent === 'function') {
        selectEvent(d)
      }

      svg
        .transition()
        .duration(750)
        .call(
          zoom.transform as any,
          d3.zoomIdentity
            .translate(width / 2, 0)
            .scale(5)
            .translate(-x(day(d.StartTime)), 0)
        )
    })
    .on('mouseover', function (d) {
      let [x, y] = d3.mouse(this)

      tooltip.html(genTooltipContent(d))
      const { width } = tooltip.node()!.getBoundingClientRect()

      y += 50
      if (x > (root.offsetWidth / 3) * 2) {
        x -= width
      }
      if (y > (root.offsetHeight / 3) * 2) {
        y -= 200
      }

      tooltip
        .style('left', x + 'px')
        .style('top', y + 'px')
        .style('opacity', 1)
    })
    .on('mouseleave', function () {
      tooltip.style('opacity', 0)
    })

  function reGen() {
    const newWidth = root.offsetWidth
    width = newWidth

    svg.attr('width', width)
    x.range([margin.left, width - margin.right])
    gXAxis.call(xAxis)
    svg.selectAll('.tick text').call(wrapText, 30)
    rects.attr('x', (d) => x(day(d.StartTime))).attr('width', genRectWidth)
  }

  d3.select(window).on('resize', reGen)

  root.appendChild(legendsRoot.node()!)
  root.appendChild(tooltip.node()!)
  root.style.position = 'relative'
}
