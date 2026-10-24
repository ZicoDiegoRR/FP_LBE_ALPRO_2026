"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

const EXIT_DURATION = 220; // ms, must match the CSS transition duration below

// Lets pages trigger "fade out, THEN navigate" instead of an instant
// route swap, and instead of next/link's Link (which navigates immediately).
const AuthTransitionContext = createContext<(href: string) => void>(() => {});
export function useAuthTransition() {
  return useContext(AuthTransitionContext);
}

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const pathname = usePathname();
  const [visible, setVisible] = useState(false);

  // Replays the fade-IN whenever the actual route changes (i.e. once the
  // new page's content has arrived), not just on first mount.
  useEffect(() => {
    setVisible(false);
    const id = requestAnimationFrame(() => setVisible(true));
    return () => cancelAnimationFrame(id);
  }, [pathname]);

  function navigate(href: string) {
    setVisible(false); // fade OUT first
    setTimeout(() => router.push(href), EXIT_DURATION);
  }

  return (
    <main
      className="min-h-screen w-full flex items-center justify-center bg-cover bg-center bg-neutral-950 text-white py-12"
      style={{ backgroundImage: "url('/img/login_bg.png')" }}
    >
      <div className="w-full max-w-xl text-center px-6">
        <h1 className="text-5xl md:text-6xl font-extrabold leading-tight">
          Borrow at Ease with PMO.
        </h1>
        <p className="mt-3 text-xs md:text-sm font-semibold tracking-[0.35em] uppercase">
          ( Peminjaman Menjadi Online )
        </p>

        <AuthTransitionContext.Provider value={navigate}>
          <div
            className={
              "mt-10 transition-all duration-200 ease-out " +
              (visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-3")
            }
          >
            {children}
          </div>
        </AuthTransitionContext.Provider>
      </div>
    </main>
  );
}
