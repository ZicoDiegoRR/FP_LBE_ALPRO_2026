"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { API_BASE_URL } from "@/lib/api";
import { useAuthTransition } from "../layout";

export default function SignupPage() {
  const router = useRouter();
  const goTo = useAuthTransition();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [agreed, setAgreed] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      const res = await fetch(`${API_BASE_URL}/signup`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ username, email, password, confirmPassword }),
      });
      const data = await res.json();

      if (!data.success) {
        setError(data.message || "Signup failed");
        return;
      }

      router.push("/login");
    } catch {
      setError("Could not reach the server.");
    } finally {
      setLoading(false);
    }
  }

  const inputClass =
    "w-full px-6 py-4 rounded-full text-gray-700 placeholder-gray-400 outline-none focus:ring-2 focus:ring-white/60";

  return (
    <>
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          type="text"
          placeholder="Username"
          required
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          className={inputClass}
        />
        <input
          type="email"
          placeholder="Email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className={inputClass}
        />
        <input
          type="password"
          placeholder="Password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className={inputClass}
        />
        <input
          type="password"
          placeholder="Confirm password"
          required
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          className={inputClass}
        />

        <div className="flex items-center justify-between gap-4 pt-1">
          <label className="flex items-start gap-2 text-left text-xs text-white/90">
            <input
              type="checkbox"
              required
              checked={agreed}
              onChange={(e) => setAgreed(e.target.checked)}
              className="mt-0.5 w-4 h-4 flex-shrink-0"
            />
            <span>
              I have read and agreed to the privacy policy, terms of service, and community guidelines.
            </span>
          </label>

          <button
            type="submit"
            disabled={loading}
            aria-label="Sign up"
            className="w-14 h-14 flex-shrink-0 rounded-full bg-white text-gray-600 flex items-center justify-center text-xl disabled:opacity-50"
          >
            {loading ? "…" : "→"}
          </button>
        </div>
      </form>

      {error && <p className="mt-4 text-sm text-red-300">{error}</p>}

      <p className="mt-6 text-sm">
        Already have an account?{" "}
        <button
          type="button"
          onClick={() => goTo("/login")}
          className="underline font-medium"
        >
          Log in
        </button>
      </p>
    </>
  );
}
