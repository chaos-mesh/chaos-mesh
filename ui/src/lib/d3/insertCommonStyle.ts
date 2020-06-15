export default function insertCommonStyle() {
  document.head.insertAdjacentHTML(
    'beforeend',
    `<style>
      text {
        fill: rgba(0, 0, 0, 0.72);
      }

      .axis path,
      .axis line {
        stroke: rgba(0, 0, 0, 0.36);
      }
    </style>`
  )
}
