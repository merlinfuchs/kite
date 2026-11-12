// Kite's logo in a 40x40 viewBox, centered at 21.5, 20.
const LOGO = `<path fill-rule="evenodd" d="M20 32C26.38 32 31.5971 27.0211 31.9778 20.7369L26.1342 21.9891C25.0184 22.7912 23.7999 23.5796 22.5 24.3301C18.4665 26.6588 14.5086 28.1269 11.5709 28.541C9.28319 28.8635 7.61411 28.5469 7.00964 27.5C6.40549 26.4535 6.96493 24.8509 8.38641 23.0322C8.72758 24.3426 9.28467 25.5658 10.0174 26.6616C10.5929 26.6593 11.3158 26.5828 12.1812 26.4057C14.7223 25.8857 18.0472 24.5915 21.5 22.598C22.4738 22.0358 23.3942 21.4547 24.2527 20.8664L27.4211 10.5692C25.379 8.96008 22.8016 8 20 8C13.4735 8 8.16384 13.2103 8.00375 19.6983C7.71232 19.9945 7.43613 20.2896 7.17616 20.5827C6.12515 21.768 5.24153 23.0305 4.73549 24.2983C4.24275 25.5329 3.95223 27.2044 4.84457 28.75C5.73692 30.2955 7.32977 30.8797 8.64525 31.0702C9.99624 31.2659 11.5315 31.132 13.0834 30.8144C13.4663 30.736 13.8589 30.6447 14.2599 30.5407C15.965 31.4712 17.9208 32 20 32Z"/><path d="M39.0626 17.1734L38.189 9.31109L30.1747 8.42061L26.6641 19.8302L39.0626 17.1734Z"/>`;
// sparkle is a four-pointed star at x, y.
const sparkle = (x: number, y: number, r: number, c = "#FDE68A") =>
  `<path d="M${x} ${y - r}Q${x + r * 0.18} ${y - r * 0.18} ${x + r} ${y}Q${
    x + r * 0.18
  } ${y + r * 0.18} ${x} ${y + r}Q${x - r * 0.18} ${y + r * 0.18} ${
    x - r
  } ${y}Q${x - r * 0.18} ${y - r * 0.18} ${x} ${y - r}z" fill="${c}"/>`;
const lg = (
  id: string,
  x1: number,
  y1: number,
  x2: number,
  y2: number,
  stops: [number, string][]
) =>
  `<linearGradient id="${id}" x1="${x1}" y1="${y1}" x2="${x2}" y2="${y2}" gradientUnits="userSpaceOnUse">${stops
    .map(([o, c]) => `<stop offset="${o}" stop-color="${c}"/>`)
    .join("")}</linearGradient>`;

export type CreatorTier = 10 | 100 | 1000;

// The SVGs of the badges, each the inside of a 24x24 viewBox. g is a prefix
// that keeps the IDs of their gradients apart when several are on a page.
export const badgeIcons = {
  // A puzzle piece: they helped put Kite together.
  contributor: (g: string) => `<defs>${lg(g + "a", 5, 4, 19, 21, [
    [0, "#7DD3FC"],
    [0.5, "#6366F1"],
    [1, "#4338CA"],
  ])}</defs>
    <g transform="translate(.6 1.2)">
      <path d="M4 8.6Q4 7 5.6 7H8.6A2.5 2.5 0 1 1 12.6 7H15.4Q17 7 17 8.6V11.4A2.5 2.5 0 1 1 17 15.4V18.4Q17 20 15.4 20H5.6Q4 20 4 18.4z" fill="url(#${g}a)"/>
      <path d="M17 15.4V18.4Q17 20 15.4 20H5.6Q4 20 4 18.4V17C8 17.8 13 17 17 15.4z" fill="#312E81" fill-opacity=".35"/>
      <circle cx="7.4" cy="10.2" r="1.3" fill="#fff" fill-opacity=".55"/>
    </g>
    ${sparkle(20.2, 4.2, 2)}`,
  // A heart with sparkles: they supported Kite early.
  early_supporter: (g: string) => `<defs>${lg(g + "a", 4, 5, 19, 20, [
    [0, "#FDA4CF"],
    [0.45, "#F43F8E"],
    [1, "#BE185D"],
  ])}</defs>
    <path d="M12 20.8C8.6 18.6 2.6 14.6 2.6 9.4 2.6 6.6 4.8 4.6 7.3 4.6c2 0 3.6 1.1 4.7 2.7 1.1-1.6 2.7-2.7 4.7-2.7 2.5 0 4.7 2 4.7 4.8 0 5.2-6 9.2-9.4 11.4z" fill="url(#${g}a)"/>
    <path d="M21.4 9.4c0 5.2-6 9.2-9.4 11.4-1.8-1.2-4.3-2.8-6.2-4.9 5.6 1.4 12.4-1.3 15.4-8.4.1.6.2 1.2.2 1.9z" fill="#831843" fill-opacity=".3"/>
    <ellipse cx="6.8" cy="8.6" rx="1.6" ry="1.1" transform="rotate(-35 6.8 8.6)" fill="#fff" fill-opacity=".6"/>
    ${sparkle(20.4, 3.2, 1.9)}${sparkle(3.2, 17.6, 1.3, "#fff")}`,
  // A planet with a ring, like Kite's logo: they pay for a plan.
  subscriber: (g: string) => `<defs>${lg(g + "a", 7, 6, 17, 18, [
    [0, "#FDBA74"],
    [0.5, "#F97316"],
    [1, "#DB2777"],
  ])}${lg(g + "r", 2, 14, 22, 9, [
    [0, "#FDE68A"],
    [1, "#F59E0B"],
  ])}
      <clipPath id="${g}c"><path d="M0 0H24V11.2L0 15.6z"/></clipPath></defs>
    <ellipse cx="12" cy="12.6" rx="10.4" ry="3.4" transform="rotate(-14 12 12.6)" fill="none" stroke="url(#${g}r)" stroke-width="1.8"/>
    <circle cx="12" cy="12" r="6.8" fill="url(#${g}a)" clip-path="url(#${g}c)"/>
    <path d="M5.4 13.6a6.8 6.8 0 0 0 13.3-1.4z" fill="url(#${g}a)"/>
    <path d="M5.6 14.4a6.8 6.8 0 0 0 12.9-1.2C15 15.4 9.6 16 5.6 14.4z" fill="#7C2D12" fill-opacity=".25"/>
    <circle cx="9.4" cy="8.8" r="1.4" fill="#fff" fill-opacity=".55"/>
    ${sparkle(20.2, 4, 1.9)}`,
  // A rocket: they were there at launch.
  early_creator: (g: string) => `<defs>${lg(g + "a", 9, 2, 15, 15, [
    [0, "#F5F3FF"],
    [1, "#A78BFA"],
  ])}${lg(g + "f", 12, 15, 12, 22, [
    [0, "#FDE047"],
    [1, "#F97316"],
  ])}</defs>
    <path d="M10 15.2Q12 23 14 15.2z" fill="url(#${g}f)"/>
    <path d="M8.7 10.6 5.6 14.2Q5.2 15.6 6.6 15.4L9 14.8zM15.3 10.6l3.1 3.6q.4 1.4-1 1.2L15 14.8z" fill="#EC4899"/>
    <path d="M12 1.8C15.4 4.2 16.4 9 15.3 15H8.7C7.6 9 8.6 4.2 12 1.8z" fill="url(#${g}a)"/>
    <path d="M12 1.8C15.4 4.2 16.4 9 15.3 15H12.6C14 10.6 14.2 5.4 12 1.8z" fill="#5B21B6" fill-opacity=".25"/>
    <circle cx="12" cy="8.4" r="2.3" fill="#4C1D95"/><circle cx="12" cy="8.4" r="1.5" fill="#38BDF8"/><circle cx="11.5" cy="7.9" r=".5" fill="#fff"/>
    ${sparkle(19.6, 5, 1.8)}${sparkle(4.4, 6.8, 1.2, "#fff")}`,
  // A kite with a tail, its color by milestone.
  creator: (g: string, tier: CreatorTier) => {
    const c = {
      10: ["#A5F3FC", "#22D3EE", "#0E7490"],
      100: ["#DDD6FE", "#8B5CF6", "#5B21B6"],
      1000: ["#FEF08A", "#FACC15", "#B45309"],
    }[tier];
    return `<defs>${lg(g + "a", 6, 2, 18, 17, [
      [0, c[0]],
      [0.5, c[1]],
      [1, c[2]],
    ])}</defs>
    <path d="M12 16.6c-1.6 1.2.8 2.3-.2 3.6-.6.8-1.6.9-1.8 2" fill="none" stroke="${
      c[2]
    }" stroke-width="1.5" stroke-linecap="round"/>
    <path d="M11.4 18.8 9.4 18l.6 1.9zM12.3 20.5l2.1-.3-1 1.8z" fill="#F472B6"/>
    <path d="M12 1.4q.8 0 1.4.6L19.6 8q.9.9 0 1.8L13.4 16.4q-1.4 1.3-2.8 0L4.4 9.8q-.9-.9 0-1.8L10.6 2q.6-.6 1.4-.6z" fill="url(#${g}a)"/>
    <path d="M12 8.9H20q0 .5-.4.9L13.4 16.4q-.7.6-1.4.7z" fill="${
      c[2]
    }" fill-opacity=".35"/>
    <path d="M12 1.4V8.9H4.1q.1-.5.3-.9L10.6 2q.6-.6 1.4-.6z" fill="#fff" fill-opacity=".3"/>
    ${
      tier === 1000
        ? sparkle(20, 2.8, 1.9) + sparkle(4, 15.6, 1.3, "#fff")
        : tier === 100
        ? sparkle(20, 2.8, 1.6)
        : ""
    }`;
  },
  // Kite's logo on an orange star seal.
  staff_pick: (
    g: string
  ) => `<defs><radialGradient id="${g}a" cx="9" cy="8" r="13" gradientUnits="userSpaceOnUse"><stop offset="0" stop-color="#FDBA74"/><stop offset=".55" stop-color="#F97316"/><stop offset="1" stop-color="#C2410C"/></radialGradient></defs>
    <g transform="translate(0 1.8)"><path d="M10.98 1.58Q12.00 0.60 13.02 1.58Q14.04 2.57 15.42 2.23Q16.80 1.89 17.19 3.25Q17.59 4.61 18.95 5.01Q20.31 5.40 19.97 6.78Q19.63 8.16 20.62 9.18Q21.60 10.20 20.62 11.22Q19.63 12.24 19.97 13.62Q20.31 15.00 18.95 15.39Q17.59 15.79 17.19 17.15Q16.80 18.51 15.42 18.17Q14.04 17.83 13.02 18.82Q12.00 19.80 10.98 18.82Q9.96 17.83 8.58 18.17Q7.20 18.51 6.81 17.15Q6.41 15.79 5.05 15.39Q3.69 15.00 4.03 13.62Q4.37 12.24 3.38 11.22Q2.40 10.20 3.38 9.18Q4.37 8.16 4.03 6.78Q3.69 5.40 5.05 5.01Q6.41 4.61 6.81 3.25Q7.20 1.89 8.58 2.23Q9.96 2.57 10.98 1.58z" fill="url(#${g}a)"/>
    <g fill="#fff" transform="translate(${12 - 21.5 * 0.33} ${
    10.4 - 20 * 0.33
  }) scale(.33)">${LOGO}</g></g>
    ${sparkle(20.8, 2.6, 1.8)}`,
};
