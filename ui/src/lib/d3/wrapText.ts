import * as d3 from 'd3'

export default function wrap(texts: d3.Selection<d3.BaseType, unknown, SVGElement, unknown>, width: number) {
  texts.each(function () {
    const text = d3.select(this)
    const words = text.text().split(/\s+/).reverse()
    let word
    let line: string[] = []
    let lineNumber = 0
    const lineHeight = 1.1
    const y = text.attr('y')
    const dy = parseFloat(text.attr('dy'))
    let tspan = text
      .text(null)
      .append('tspan')
      .attr('x', 0)
      .attr('y', y)
      .attr('dy', dy + 'em')

    while ((word = words.pop())) {
      line.push(word)
      tspan.text(line.join(' '))

      if (tspan.node()!.getComputedTextLength() > width) {
        line.pop()
        tspan.text(line.join(' '))
        line = [word]
        tspan = text
          .append('tspan')
          .attr('x', 0)
          .attr('y', y)
          .attr('dy', `${++lineNumber * lineHeight + dy}em`)
          .text(word)
      }
    }
  })
}
