export default function insertCommonStyle() {
  document.head.insertAdjacentHTML(
    'beforeend',
    `<style>
      .chaos-chart text {
        fill: rgba(0, 0, 0, 0.87);
      }

      .chaos-chart .axis path,
      .chaos-chart .axis line {
        stroke: rgba(0, 0, 0, 0.38);
      }

      .chaos-events-legends {
        position: absolute;
        top: 0;
        right: 1rem;
        display: flex;
        flex-direction: column;
        align-items: end;
      }

      .chaos-events-legends > div {
        display: flex;
        align-items: center;
      }
    </style>`
  )
}
