export type CreatorTier = 10 | 100 | 1000;

// The SVGs of the badges, each the inside of a 24x24 viewBox. g is a prefix
// that keeps the IDs of their gradients apart when several are on a page.
export const badgeIcons = {
  contributor: (g: string) => `
    <defs>
      <linearGradient id="${g}a" x1="3" y1="2" x2="21" y2="22" gradientUnits="userSpaceOnUse"><stop stop-color="#818CF8"/><stop offset="1" stop-color="#7C3AED"/></linearGradient>
      <linearGradient id="${g}s" x1="12" y1="2" x2="12" y2="12" gradientUnits="userSpaceOnUse"><stop stop-color="#fff" stop-opacity=".45"/><stop offset="1" stop-color="#fff" stop-opacity="0"/></linearGradient>
    </defs>
    <path d="M12 1.8 20.8 6.9v10.2L12 22.2 3.2 17.1V6.9z" fill="url(#${g}a)" stroke="#4C1D95" stroke-opacity=".35" stroke-width=".8" stroke-linejoin="round"/>
    <path d="M12 2.6 20 7.2v4.3C15 9.6 9 9.6 4 11.5V7.2z" fill="url(#${g}s)"/>
    <path d="m9.2 9.2-2.6 2.8 2.6 2.8M14.8 9.2l2.6 2.8-2.6 2.8M12.9 8.4l-1.8 7.2" fill="none" stroke="#fff" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round"/>`,
  early_supporter: (g: string) => `
    <defs>
      <linearGradient id="${g}a" x1="3" y1="4" x2="20" y2="21" gradientUnits="userSpaceOnUse"><stop stop-color="#FB7185"/><stop offset="1" stop-color="#DB2777"/></linearGradient>
    </defs>
    <path d="M12 20.6s-8.4-4.8-8.4-10.8A4.6 4.6 0 0 1 12 7.3a4.6 4.6 0 0 1 8.4 2.5c0 6-8.4 10.8-8.4 10.8z" fill="url(#${g}a)" stroke="#9D174D" stroke-opacity=".35" stroke-width=".8" stroke-linejoin="round"/>
    <path d="M7.6 7.4c-1.6.2-2.6 1.4-2.6 2.9" fill="none" stroke="#fff" stroke-opacity=".7" stroke-width="1.3" stroke-linecap="round"/>
    <path d="M19 1.6l.7 1.7 1.7.7-1.7.7-.7 1.7-.7-1.7-1.7-.7 1.7-.7z" fill="#FDE68A"/>`,
  subscriber: (g: string) => `
    <defs>
      <linearGradient id="${g}a" x1="4" y1="4" x2="20" y2="21" gradientUnits="userSpaceOnUse"><stop stop-color="#FDBA74"/><stop offset=".5" stop-color="#F97316"/><stop offset="1" stop-color="#C2410C"/></linearGradient>
    </defs>
    <path d="M7.2 3.6h9.6l4.4 5.6L12 21 2.8 9.2z" fill="url(#${g}a)" stroke="#7C2D12" stroke-opacity=".35" stroke-width=".8" stroke-linejoin="round"/>
    <path d="M2.8 9.2h18.4M7.2 3.6 9 9.2l3 11.8 3-11.8 1.8-5.6M9 9.2 12 3.6l3 5.6" fill="none" stroke="#fff" stroke-opacity=".55" stroke-width=".9" stroke-linejoin="round"/>
    <path d="M7.2 3.6h4.8L9 9.2H2.8z" fill="#fff" fill-opacity=".35"/>`,
  early_creator: (g: string) => `
    <defs>
      <linearGradient id="${g}a" x1="12" y1="4" x2="12" y2="16" gradientUnits="userSpaceOnUse"><stop stop-color="#FDE047"/><stop offset="1" stop-color="#F97316"/></linearGradient>
      <linearGradient id="${g}b" x1="2" y1="15" x2="22" y2="22" gradientUnits="userSpaceOnUse"><stop stop-color="#F472B6"/><stop offset="1" stop-color="#A855F7"/></linearGradient>
    </defs>
    <path d="M12 2.4v2.4M4.2 6.2l1.7 1.7M19.8 6.2l-1.7 1.7M1.8 13.6h2.4M19.8 13.6h2.4" stroke="#FBBF24" stroke-width="1.8" stroke-linecap="round"/>
    <path d="M5.6 15a6.4 6.4 0 0 1 12.8 0z" fill="url(#${g}a)"/>
    <rect x="2" y="15" width="20" height="6.6" rx="3.3" fill="url(#${g}b)"/>
    <path d="M6 18.3h5M13.5 18.3h4" stroke="#fff" stroke-opacity=".6" stroke-width="1.3" stroke-linecap="round"/>`,
  // A kite with a tail, in bronze, blue or gold by milestone.
  creator: (g: string, tier: CreatorTier) => {
    const c = {
      10: ["#FDBA74", "#B45309", "#78350F"],
      100: ["#E0F2FE", "#60A5FA", "#1E3A8A"],
      1000: ["#FEF08A", "#EAB308", "#713F12"],
    }[tier];
    return `
    <defs>
      <linearGradient id="${g}a" x1="4" y1="1" x2="20" y2="17" gradientUnits="userSpaceOnUse"><stop stop-color="${c[0]}"/><stop offset="1" stop-color="${c[1]}"/></linearGradient>
    </defs>
    <path d="M12.8 16.4c-.8 1.6.9 2.3.1 3.6-.7 1.1-2 .9-2.4 2.2" fill="none" stroke="${c[1]}" stroke-width="1.6" stroke-linecap="round"/>
    <path d="m14.4 18.6 2-.4-.6 1.8zM10.6 21.2l-1.9.4 1.4 1.2z" fill="${c[1]}"/>
    <path d="M12 .9 20.6 8.4 12 17.2 3.4 8.4z" fill="url(#${g}a)" stroke="${c[2]}" stroke-opacity=".45" stroke-width=".8" stroke-linejoin="round"/>
    <path d="M12 .9v16.3M3.4 8.4h17.2" stroke="${c[2]}" stroke-opacity=".4" stroke-width=".9"/>
    <path d="M12 .9 3.4 8.4H12z" fill="#fff" fill-opacity=".4"/>`;
  },
  staff_pick: (g: string) => `
    <defs>
      <linearGradient id="${g}a" x1="3" y1="2" x2="21" y2="20" gradientUnits="userSpaceOnUse"><stop stop-color="#FB923C"/><stop offset="1" stop-color="#EA580C"/></linearGradient>
    </defs>
    <path d="M8.2 15.5 6.4 22l3.1-1.6 2.5 2.1 1.2-6.6zM15.8 15.5l1.8 6.5-3.1-1.6-2.5 2.1-1.2-6.6z" fill="#BE123C"/>
    <path d="m12 1.4 2.1 1.3 2.4-.4 1.3 2.1 2.2 1.1-.1 2.5 1.3 2.1-1.3 2.1.1 2.5-2.2 1.1-1.3 2.1-2.4-.4-2.1 1.3-2.1-1.3-2.4.4-1.3-2.1-2.2-1.1.1-2.5L3.2 10l1.3-2.1-.1-2.5 2.2-1.1 1.3-2.1 2.4.4z" fill="url(#${g}a)" stroke="#9A3412" stroke-opacity=".35" stroke-width=".8" stroke-linejoin="round"/>
    <g transform="translate(4.7 5.2) scale(.34)">
      <path fill-rule="evenodd" clip-rule="evenodd" d="M20 32C26.38 32 31.5971 27.0211 31.9778 20.7369L26.1342 21.9891C25.0184 22.7912 23.7999 23.5796 22.5 24.3301C18.4665 26.6588 14.5086 28.1269 11.5709 28.541C9.28319 28.8635 7.61411 28.5469 7.00964 27.5C6.40549 26.4535 6.96493 24.8509 8.38641 23.0322C8.72758 24.3426 9.28467 25.5658 10.0174 26.6616C10.5929 26.6593 11.3158 26.5828 12.1812 26.4057C14.7223 25.8857 18.0472 24.5915 21.5 22.598C22.4738 22.0358 23.3942 21.4547 24.2527 20.8664L27.4211 10.5692C25.379 8.96008 22.8016 8 20 8C13.4735 8 8.16384 13.2103 8.00375 19.6983C7.71232 19.9945 7.43613 20.2896 7.17616 20.5827C6.12515 21.768 5.24153 23.0305 4.73549 24.2983C4.24275 25.5329 3.95223 27.2044 4.84457 28.75C5.73692 30.2955 7.32977 30.8797 8.64525 31.0702C9.99624 31.2659 11.5315 31.132 13.0834 30.8144C13.4663 30.736 13.8589 30.6447 14.2599 30.5407C15.965 31.4712 17.9208 32 20 32Z" fill="#fff"/>
      <path d="M39.0626 17.1734L38.189 9.31109L30.1747 8.42061L26.6641 19.8302L39.0626 17.1734Z" fill="#fff"/>
    </g>`,
};
