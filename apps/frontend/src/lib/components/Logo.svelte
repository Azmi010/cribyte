<script lang="ts">
  let { class: className = "size-6" }: { class?: string } = $props();
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  viewBox="0 0 512 512"
  class={className}
  fill="none"
>
  <defs>
    <linearGradient id="amberGradTrans" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#F5D061" />
      <stop offset="50%" stop-color="#E6B93A" />
      <stop offset="100%" stop-color="#C29013" />
    </linearGradient>
    <linearGradient id="cubeTopTrans" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#FFE58F" />
      <stop offset="50%" stop-color="#E6B93A" />
    </linearGradient>
    <linearGradient id="cubeLeftTrans" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#D6A419" />
      <stop offset="100%" stop-color="#9E7307" />
    </linearGradient>
    <linearGradient id="cubeRightTrans" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#E6B93A" />
      <stop offset="100%" stop-color="#B8850E" />
    </linearGradient>
  </defs>

  <g transform="translate(0, 0)">
    <!-- Outer bracket / Shielding Cloud curve contour (Stylized 'C' for CriByte) -->
    <path
      d="M 330 148 C 240 125 155 180 142 270 C 130 355 192 418 290 412 C 342 409 375 385 390 365"
      fill="none"
      stroke="url(#amberGradTrans)"
      stroke-width="18"
      stroke-linecap="round"
      stroke-linejoin="round"
    />

    <!-- Data Byte nodes on the orbit/bracket -->
    <circle cx="330" cy="148" r="9" fill="#FFE58F" />
    <circle cx="142" cy="270" r="8" fill="#E6B93A" />
    <circle cx="390" cy="365" r="9" fill="#FFE58F" />

    <!-- Central Isometric Byte Cube (Storage block unit) -->
    <!-- Top Face -->
    <polygon points="268,175 338,215 268,255 198,215" fill="url(#cubeTopTrans)" />
    <!-- Left Face -->
    <polygon points="198,215 268,255 268,335 198,295" fill="url(#cubeLeftTrans)" />
    <!-- Right Face -->
    <polygon points="268,255 338,215 338,295 268,335" fill="url(#cubeRightTrans)" />

    <!-- Inner Core Cutout / Data slot in isometric cube -->
    <line
      x1="268"
      y1="205"
      x2="268"
      y2="225"
      stroke="#1C1E21"
      stroke-width="4"
      stroke-linecap="round"
      opacity="0.45"
    />
    <line
      x1="233"
      y1="275"
      x2="233"
      y2="255"
      stroke="#FFE58F"
      stroke-width="3"
      stroke-linecap="round"
      opacity="0.75"
    />
    <line
      x1="303"
      y1="275"
      x2="303"
      y2="255"
      stroke="#FFE58F"
      stroke-width="3"
      stroke-linecap="round"
      opacity="0.75"
    />

    <!-- Central terminal point indicator -->
    <circle cx="268" cy="255" r="5" fill="#FFFFFF" />
  </g>
</svg>

