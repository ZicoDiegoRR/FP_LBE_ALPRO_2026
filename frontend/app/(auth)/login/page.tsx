"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { API_BASE_URL } from "@/lib/api";
import { useAuthTransition } from "../layout";

export default function LoginPage() {
  const router = useRouter();
  const goTo = useAuthTransition();
  const [usernameOrEmail, setUsernameOrEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      const res = await fetch(`${API_BASE_URL}/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ usernameOrEmail, password }),
      });
      const data = await res.json();

      if (!data.success) {
        setError(data.message || "Login failed");
        return;
      }

      router.push("/dashboard");
    } catch {
      setError("Could not reach the server.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          type="text"
          placeholder="Username or Email"
          required
          value={usernameOrEmail}
          onChange={(e) => setUsernameOrEmail(e.target.value)}
          className="w-full px-6 py-4 rounded-full text-gray-700 placeholder-gray-400 outline-none focus:ring-2 focus:ring-white/60"
        />

        <div className="flex gap-3">
          <input
            type="password"
            placeholder="Password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="flex-1 px-6 py-4 rounded-full text-gray-700 placeholder-gray-400 outline-none focus:ring-2 focus:ring-white/60"
          />
          <button
            type="submit"
            disabled={loading}
            aria-label="Sign in"
            className="w-14 h-14 flex-shrink-0 rounded-full bg-white text-gray-600 flex items-center justify-center text-xl disabled:opacity-50"
          >
            {loading ? "…" : "→"}
          </button>
        </div>
      </form>

      {error && <p className="mt-4 text-sm text-red-300">{error}</p>}

      <p className="mt-6 text-sm">
        Or{" "}
        <button
          type="button"
          onClick={() => goTo("/signup")}
          className="underline font-medium"
        >
          sign up
        </button>{" "}
        here!
      </p>
    </>
  );
}
