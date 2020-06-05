import * as d3 from 'd3'

import { Event } from 'api/events.type'
import day from './dayjs'

const margin = {
  top: 30,
  right: 30,
  bottom: 30,
  left: 30,
}

export default function gen({
  root,
  width,
  height,
  events,
  selectEvent,
}: {
  root: HTMLElement
  width: number
  height: number
  events: Event[]
  selectEvent: (e: Event) => void
}) {
  document.head.insertAdjacentHTML(
    'beforeend',
    `<style>
      .x-axis text {
        font-weight: bold
      }
    </style>`
  )

  const svg = d3.select(root).append('svg').attr('viewBox', `0, 0, ${width}, ${height}`)

  const now = day()

  const x = d3
    .scaleLinear()
    .domain([now.subtract(1, 'h'), now])
    .range([margin.left, width - margin.right])
  const xAxis = d3.axisBottom(x)

  const gXAxis = svg
    .append('g')
    .attr('class', 'x-axis')
    .attr('transform', `translate(0, ${height - margin.bottom})`)
    .call(xAxis.tickFormat(d3.timeFormat('%H:%M') as any))

  const rects = svg
    .selectAll()
    .data(events)
    .enter()
    .append('rect')
    .attr('x', (d) => x(day(d.StartTime)))
    .attr('y', height / 3)
    .attr('width', (d) => (d.FinishTime ? x(day(d.FinishTime)) - x(day(d.StartTime)) : x(day()) - x(day(d.StartTime))))
    .attr('height', 30)
    .attr('fill', (d) => (d.FinishTime ? '#388e3c' : '#172d72'))
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
            StartTime: ${day(d.StartTime).format('YYYY-MM-DD HH:mm:ss')}
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

  svg.on('mousemove', function (d) {
    let [x, y] = d3.mouse(this)

    x += 30
    y += 45
    if (x > width / 2) {
      x -= 120
    }

    tooltip.style('left', x + 'px').style('top', y + 'px')
  })

  root.appendChild(tooltip.node()!)
}
