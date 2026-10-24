"use client";

import Link from "next/link";
import Image from "next/image";
import { usePathname, useRouter } from "next/navigation";
import { API_BASE_URL } from "@/lib/api";
import { ICONS } from "@/lib/icons";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: ICONS.dashboard },
  { href: "/borrow", label: "Borrow Now!", icon: ICONS.borrowNow },
  { href: "/profile", label: "My Profile", icon: ICONS.profile },
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();

  async function handleSignOut() {
    try {
      await fetch(`${API_BASE_URL}/logout`, {
        method: "POST",
        credentials: "include",
      });
    } finally {
      router.push("/login");
    }
  }

  return (
    <div
      className="min-h-screen w-full bg-cover bg-center bg-neutral-950 flex gap-6 p-6"
      style={{ backgroundImage: "url('/img/dashboard_bg.png')" }}
    >
      {/* Sidebar */}
      <aside className="w-64 flex-shrink-0 bg-white rounded-2xl flex flex-col justify-between p-4">
        <nav className="space-y-2">
          {navItems.map((item) => {
            const active = pathname === item.href || pathname.startsWith(item.href + "/");
            return (
              <Link
                key={item.href}
                href={item.href}
                className={
                  "flex items-center gap-3 px-4 py-3 rounded-xl font-semibold transition " +
                  (active
                    ? "bg-slate-700 text-white"
                    : "text-slate-900 hover:bg-slate-100")
                }
              >
                <Image
                  src={item.icon}
                  alt=""
                  width={20}
                  height={20}
                  className={active ? "invert" : ""}
                />
                {item.label}
              </Link>
            );
          })}
        </nav>

        <button
          onClick={handleSignOut}
          className="flex items-center gap-3 px-4 py-3 rounded-xl font-semibold text-slate-900 hover:bg-slate-100 transition"
        >
          <Image src={ICONS.signOut} alt="" width={20} height={20} />
          Sign Out
        </button>
      </aside>

      {/* Page content */}
      <div className="flex-1 min-w-0">{children}</div>
    </div>
  );
}