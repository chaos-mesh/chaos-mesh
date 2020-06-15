export default function style() {
  document.head.insertAdjacentHTML(
    'beforeend',
    `<style>
      .x-axis path,
      .x-axis line {
        stroke: rgba(0, 0, 0, 0.36);
      }

      .x-axis text {
        fill: rgba(0, 0, 0, 0.72);
      }
    </style>`
  )
}
