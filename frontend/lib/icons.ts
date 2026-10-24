// Centralized icon paths. Point these at files in /public (e.g. "/icons/x.svg")
// or a hosted URL, same pattern as the login background/default pfp.
export const ICONS = {
  // Sidebar nav
  dashboard: "/icons/dashboard.png",
  borrowNow: "/icons/borrow-now.png",
  profile: "/icons/profile.png",
  signOut: "/icons/sign-out.png",

  // Stat cards (the large faded icon behind each number)
  statTotal: "/icons/checklist.svg",
  statOngoing: "/icons/gear.svg",
  statWaiting: "/icons/clock.svg",
  statCancelledRejected: "/icons/x-circle.svg",

  // Recent Activity row icons, keyed by ActivityItem.statusType from the API.
  // clock/x-circle/check-circle are reused from the stat cards above —
  // same file, just rendered small + solid instead of large + faded.
  activity: {
    cancelled: "/icons/x-circle.svg",
    rejected: "/icons/x-circle.svg",
    failed: "/icons/x-circle.svg",
    pending: "/icons/clock.svg",
    accepted: "/icons/check-circle.svg",
    finished: "/icons/check-circle.svg",
    ongoing: "/icons/refresh.svg",
    none: "",
  },
} as const;