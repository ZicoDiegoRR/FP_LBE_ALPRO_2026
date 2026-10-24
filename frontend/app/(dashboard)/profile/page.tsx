"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { API_BASE_URL } from "@/lib/api";
import { ICONS } from "@/lib/icons";

// Rectangular card background — a separate fixed image, distinct from the
// user's actual profile picture (user.pfpUrl).
const PROFILE_BANNER_BACKGROUND = "/img/profile_banner.png";

type BorrowRequest = {
  id: string;
  title: string;
  bookedAt: string;
  statusMessage: string;
  statusMessageUpdateTime: string;
  statusType: keyof typeof ICONS.activity;
};

type ProfileData = {
  user: {
    id: string;
    name: string;
    email: string;
    pfpUrl: string;
    roleName: string;
    creationDate: string;
  };
  borrowRequests: BorrowRequest[];
};

export default function ProfilePage() {
  const router = useRouter();
  const [data, setData] = useState<ProfileData | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const res = await fetch(`${API_BASE_URL}/profile`, {
          credentials: "include",
        });
        const json = await res.json();
        if (json.success) setData(json);
        else console.error("Failed to load profile:", json.message);
      } catch (err) {
        console.error("Could not reach /profile:", err);
      }
    }
    load();
  }, []);

  function handleRequestClick(id: string) {
    router.push(`/profile/${id}`);
  }

  if (!data) {
    return <div className="text-white p-8">Loading...</div>;
  }

  const { user, borrowRequests } = data;

  return (
    <div className="space-y-6">
      {/* Profile banner */}
      <div
        className="relative rounded-2xl overflow-hidden bg-slate-200 bg-cover bg-center p-8 flex items-center gap-6"
        style={{ backgroundImage: `url('${PROFILE_BANNER_BACKGROUND}')` }}
      >
        <div className="relative flex-shrink-0">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={user.pfpUrl}
            alt={user.name}
            className="w-28 h-28 rounded-full object-cover bg-white"
          />
          <button
            className="absolute bottom-0 right-0 w-8 h-8 rounded-full bg-slate-900 flex items-center justify-center text-white text-sm"
            aria-label="Edit profile picture"
            title="Editing not available yet"
          >
            ✎
          </button>
        </div>

        <div>
          <h1 className="text-2xl font-bold text-slate-900">
            {user.name} — {user.roleName}
          </h1>
          <p className="text-slate-600 text-sm mt-1">{user.id}</p>
          <p className="text-slate-600 text-sm">{user.email}</p>
          <p className="text-slate-600 text-sm">
            Created at: {user.creationDate}
          </p>
        </div>
      </div>

      {/* Borrow requests */}
      <div className="bg-white rounded-2xl p-6">
        <div className="flex items-baseline justify-between mb-4">
          <h2 className="text-2xl font-bold">Your Borrow Requests</h2>
        </div>

        <div className="max-h-[500px] overflow-y-auto divide-y divide-slate-100">
          {borrowRequests.map((req) => (
            <button
              key={req.id}
              onClick={() => handleRequestClick(req.id)}
              className="w-full flex items-start gap-4 py-4 text-left hover:bg-slate-50 transition rounded-lg px-2 -mx-2"
            >
              {ICONS.activity[req.statusType] && (
                <Image
                  src={ICONS.activity[req.statusType]}
                  alt=""
                  width={28}
                  height={28}
                  className="flex-shrink-0 mt-0.5"
                />
              )}

              <div className="flex-1 min-w-0">
                <p className="font-semibold text-slate-900">{req.title}</p>
                <p className="text-slate-500 text-sm">
                  Created at: {req.bookedAt}
                </p>
                <p className="text-slate-500 text-sm">
                  Status message: {req.statusMessage}
                </p>
              </div>

              <div className="text-right flex-shrink-0">
                <p className="text-sm text-slate-500">Last updated:</p>
                <p className="text-sm text-slate-500">
                  {req.statusMessageUpdateTime}
                </p>
              </div>
            </button>
          ))}

          {borrowRequests.length === 0 && (
            <p className="text-slate-400 py-8 text-center">
              You haven't made any borrow requests yet.
            </p>
          )}
        </div>
      </div>
    </div>
  );
}