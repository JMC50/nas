<script lang="ts">
  interface Props {
    points: number[];
    color: string;
    height?: number;
    max?: number;
  }

  let { points, color, height = 64, max = 100 }: Props = $props();

  const WIDTH = 320;

  const series = $derived(points.length >= 2 ? points : [...points, ...points, 0]);
  const lineD = $derived(buildLine(series, WIDTH, height, max));
  const areaD = $derived(buildArea(series, WIDTH, height, max));
  const gradId = $derived(`spark-${color.replace(/[^a-zA-Z0-9]/g, "")}`);

  function pointAt(index: number, total: number, value: number, width: number, top: number): [number, number] {
    const x = total <= 1 ? width : (index / (total - 1)) * width;
    const ratio = top === 0 ? 0 : Math.min(Math.max(value, 0), top) / top;
    const y = height - ratio * height;
    return [x, y];
  }

  function buildLine(values: number[], width: number, top: number, ceiling: number): string {
    if (values.length === 0) return "";
    return values
      .map((value, index) => {
        const [x, y] = pointAt(index, values.length, value, width, ceiling);
        return `${index === 0 ? "M" : "L"}${x.toFixed(2)},${y.toFixed(2)}`;
      })
      .join(" ");
  }

  function buildArea(values: number[], width: number, top: number, ceiling: number): string {
    if (values.length === 0) return "";
    const line = buildLine(values, width, top, ceiling);
    return `${line} L${width.toFixed(2)},${height.toFixed(2)} L0,${height.toFixed(2)} Z`;
  }
</script>

<svg
  viewBox="0 0 {WIDTH} {height}"
  preserveAspectRatio="none"
  class="w-full block"
  style="height: {height}px;"
  aria-hidden="true"
>
  <defs>
    <linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
      <stop offset="0%" stop-color={color} stop-opacity="0.35" />
      <stop offset="100%" stop-color={color} stop-opacity="0" />
    </linearGradient>
  </defs>
  <path d={areaD} fill="url(#{gradId})" />
  <path
    d={lineD}
    fill="none"
    stroke={color}
    stroke-width="1.75"
    stroke-linejoin="round"
    stroke-linecap="round"
  />
</svg>
