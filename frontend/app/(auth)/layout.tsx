"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

const EXIT_DURATION = 220; // ms, 

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
      className="relative flex min-h-screen w-full items-center justify-center overflow-hidden bg-neutral-950 bg-cover bg-center px-4 py-6 text-white"
      style={{ backgroundImage: "url('/img/big.png')" }}
    >
    
      <div className="absolute inset-0 bg-gradient-to-br from-neutral-950/90 via-neutral-950/75 to-indigo-950/70 backdrop-blur-[2px]" />

      <div className="relative z-10 w-full max-w-sm">
       
        <div className="rounded-2xl border border-white/10 bg-white/[0.05] p-6 shadow-2xl shadow-black/40 backdrop-blur-xl">
          
          <div className="text-center">
            <div className="mx-auto mb-3 flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-indigo-500 to-violet-600 shadow-lg shadow-indigo-500/30">
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.8"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="h-5 w-5 text-white"
                aria-hidden
              >
                <path d="M4 19.5V5a2 2 0 0 1 2-2h13v16H6a2 2 0 0 0-2 2" />
                <path d="M4 19.5A2.5 2.5 0 0 0 6.5 22H19v-3" />
                <path d="M9 8h6" />
              </svg>
            </div>

            <h1 className="text-2xl font-bold leading-snug tracking-tight text-white md:text-3xl">
                Pinjam Praktis Lewat PMO
            </h1>
            <p className="mt-3 inline-block rounded-full border border-indigo-400/30 bg-indigo-400/10 px-4 py-1 text-xs font-semibold uppercase tracking-[0.2em] text-indigo-200 md:text-sm">
                Peminjaman Menjadi Online
            </p>
          </div>

          <AuthTransitionContext.Provider value={navigate}>
            <div
              className={
                "mt-5 transition-all duration-300 ease-out " +
                // hanya input teks (checkbox/radio tidak disentuh)
                "[&_input:not([type=checkbox]):not([type=radio])]:w-full " +
                "[&_input:not([type=checkbox]):not([type=radio])]:rounded-xl " +
                "[&_input:not([type=checkbox]):not([type=radio])]:border " +
                "[&_input:not([type=checkbox]):not([type=radio])]:border-white/10 " +
                "[&_input:not([type=checkbox]):not([type=radio])]:bg-white/10 " +
                "[&_input:not([type=checkbox]):not([type=radio])]:px-4 " +
                "[&_input:not([type=checkbox]):not([type=radio])]:py-2.5 " +
                "[&_input:not([type=checkbox]):not([type=radio])]:text-sm " +
                "[&_input:not([type=checkbox]):not([type=radio])]:text-white " +
                "[&_input:not([type=checkbox]):not([type=radio])]:outline-none " +
                "[&_input:not([type=checkbox]):not([type=radio])]:placeholder:text-neutral-400 " +
                "[&_input:not([type=checkbox]):not([type=radio]):focus]:border-indigo-400 " +
                "[&_input:not([type=checkbox]):not([type=radio]):focus]:ring-2 " +
                "[&_input:not([type=checkbox]):not([type=radio]):focus]:ring-indigo-400/30 " +
                // checkbox tetap kecil
                "[&_input[type=checkbox]]:h-4 [&_input[type=checkbox]]:w-4 " +
                "[&_input[type=checkbox]]:shrink-0 [&_input[type=checkbox]]:accent-indigo-500 " +
                (visible
                  ? "opacity-100 translate-y-0"
                  : "opacity-0 translate-y-3")
              }
            >
              {children}
            </div>
          </AuthTransitionContext.Provider>
        </div>
      </div>
    </main>
  );
}