"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { API_BASE_URL } from "@/lib/api";
import { ICONS } from "@/lib/icons";

type ActivityItem = {
  statusType:
    | "cancelled"
    | "pending"
    | "accepted"
    | "rejected"
    | "ongoing"
    | "finished"
    | "failed"
    | "none";
  line1: string;
  line2?: string;
};

type DashboardData = {
  user: { name: string };
  stats: {
    totalBorrowings: number;
    ongoing: number;
    waitingForApproval: number;
    cancelledRejected: number;
  };
  recentActivity: ActivityItem[];
};

function StatCard({
  title,
  value,
  icon,
}: {
  title: string;
  value: number;
  icon: string;
}) {
  return (
    <div className="relative bg-white rounded-2xl p-6 h-40 flex flex-col overflow-hidden">
      <h3 className="font-bold text-slate-900 leading-tight">{title}</h3>
      <div className="flex-1 flex items-center justify-center relative">
        <Image
          src={icon}
          alt=""
          width={80}
          height={80}
          className="absolute opacity-10"
        />
        <span className="relative text-6xl font-black text-slate-900">
          {value}
        </span>
      </div>
    </div>
  );
}

export default function DashboardPage() {
  const router = useRouter();
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const res = await fetch(`${API_BASE_URL}/dashboard`, {
          credentials: "include",
        });
        if (res.status === 401) {
          router.replace("/login");
          return;
        }
        const json = await res.json();
        setData(json);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [router]);

  if (loading) {
    return <div className="text-white p-8">Loading...</div>;
  }
  if (!data) {
    return <div className="text-white p-8">Could not load dashboard.</div>;
  }

  const { user, stats, recentActivity } = data;

  return (
    <div className="space-y-6">
      <h1 className="text-4xl font-black text-white">
        Welcome, {user.name}!
      </h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          title="Total Borrowings"
          value={stats.totalBorrowings}
          icon={ICONS.statTotal}
        />
        <StatCard
          title="Ongoing"
          value={stats.ongoing}
          icon={ICONS.statOngoing}
        />
        <StatCard
          title="Waiting for Approval"
          value={stats.waitingForApproval}
          icon={ICONS.statWaiting}
        />
        <StatCard
          title="Cancelled/ Rejected"
          value={stats.cancelledRejected}
          icon={ICONS.statCancelledRejected}
        />
      </div>

      <div className="bg-white rounded-2xl p-6">
        <h2 className="text-2xl font-bold underline mb-4">Recent Activity</h2>
        <ul className="space-y-4">
          {recentActivity.map((item, i) => {
            const icon = ICONS.activity[item.statusType];
            return (
              <li key={i} className="flex items-start gap-3">
                {icon && (
                  <Image
                    src={icon}
                    alt=""
                    width={20}
                    height={20}
                    className="mt-0.5 flex-shrink-0"
                  />
                )}
                <div>
                  <p className="text-slate-900">{item.line1}</p>
                  {item.line2 && (
                    <p className="text-slate-600 text-sm">{item.line2}</p>
                  )}
                </div>
              </li>
            );
          })}
        </ul>
      </div>
    </div>
  );
}