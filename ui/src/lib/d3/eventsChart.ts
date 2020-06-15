import * as d3 from 'd3'

import { Event } from 'api/events.type'
import day from 'lib/dayjs'
import style from './style'

const margin = {
  top: 30,
  right: 30,
  bottom: 30,
  left: 30,
}

export default function gen({
  root,
  events,
  selectEvent,
}: {
  root: HTMLElement
  events: Event[]
  selectEvent: (e: Event) => void
}) {
  style()

  const width = root.offsetWidth
  const height = root.offsetHeight

  const svg = d3.select(root).append('svg').attr('width', width).attr('height', height)

  const now = day()

  const x = d3
    .scaleLinear()
    .domain([now.subtract(1, 'h'), now])
    .range([margin.left, width - margin.right])
  const xAxis = d3
    .axisBottom(x)
    .ticks(6)
    .tickFormat(d3.timeFormat('%H:%M') as any)

  const gXAxis = svg
    .append('g')
    .attr('class', 'x-axis')
    .attr('transform', `translate(0, ${height - margin.bottom})`)
    .call(xAxis)

  const gMain = svg.append('g')

  const rects = gMain
    .selectAll()
    .data(events)
    .enter()
    .append('rect')
    .attr('x', (d) => x(day(d.StartTime)))
    .attr('y', height / 3)
    .attr('width', (d) => (d.FinishTime ? x(day(d.FinishTime)) - x(day(d.StartTime)) : x(day()) - x(day(d.StartTime))))
    .attr('height', height / 3)
    .attr('fill', '#172d72')
    .style('cursor', 'pointer')

  const zoom = d3.zoom().on('zoom', zoomd)

  svg.call(zoom as any)

  function zoomd() {
    const eventTransform = d3.event.transform

    const newX = eventTransform.rescaleX(x)

    gXAxis.call(xAxis.scale(newX))
    rects
      .attr('x', (d) => newX(day(d.StartTime)))
      .attr('width', (d) =>
        d.FinishTime ? newX(day(d.FinishTime)) - newX(day(d.StartTime)) : newX(day()) - newX(day(d.StartTime))
      )
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
  }

  function genTooltipContent(d: Event) {
    return `<b>Experiment: ${d.Experiment}</b>
            <br />
            <span style="color: rgba(0, 0, 0, 0.67);">StartTime: ${day(d.StartTime).format(
              'YYYY-MM-DD HH:mm:ss A'
            )}</span>
            <br />
            <br />
            <b>Pods:</b>
            <ul style="margin-top: 0.25rem; margin-bottom: 0;">
              ${d.Pods.map((p) => `<li><b>ip:</b> ${p.PodIP}<br /><b>name:</b> ${p.PodName}</li>`).join('')}
            </ul>
            `
  }

  rects
    .on('click', function (d) {
      selectEvent(d)
    })
    .on('mouseover', function (d) {
      tooltip.style('opacity', 1).html(genTooltipContent(d))
    })
    .on('mouseleave', function () {
      tooltip.style('opacity', 0)
    })

  svg.on('mousemove', function () {
    let [x, y] = d3.mouse(this)

    x += 50
    y += 100
    if (x > (root.offsetWidth / 3) * 2) {
      x -= 325
    }

    tooltip.style('left', x + 'px').style('top', y + 'px')
  })

  root.appendChild(tooltip.node()!)

  function reGen() {
    const newWidth = root.offsetWidth

    svg.attr('width', newWidth)
    x.range([margin.left, newWidth - margin.right])
    gXAxis.call(xAxis)
    rects
      .attr('x', (d) => x(day(d.StartTime)))
      .attr('width', (d) =>
        d.FinishTime ? x(day(d.FinishTime)) - x(day(d.StartTime)) : x(day()) - x(day(d.StartTime))
      )
  }

  d3.select(window).on('resize', reGen)
}
