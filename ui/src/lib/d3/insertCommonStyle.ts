export default function insertCommonStyle() {
  document.head.insertAdjacentHTML(
    'beforeend',
    `<style>
      .chaos-events-chart text {
        fill: rgba(0, 0, 0, 0.72);
      }

      .chaos-events-chart .axis path,
      .chaos-events-chart .axis line {
        stroke: rgba(0, 0, 0, 0.36);
      }

      .chaos-events-legends {
        position: absolute;
        top: 0;
        right: 15px;
        display: flex;
        flex-direction: column;
      }

      .chaos-events-legends > div {
        display: flex;
        justify-content: end;
        align-items: center;
      }
    </style>`
  )
}
