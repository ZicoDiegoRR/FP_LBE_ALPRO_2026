"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import Image from "next/image";
import { API_BASE_URL } from "@/lib/api";
import { ICONS } from "@/lib/icons";

type BorrowDetail = {
  id: string;
  title: string;
  bookedAt: string;
  scheduledAt: string;
  statusMessageUpdateTime: string;
  finishedAt: string | null;
  labName: string;
  labPictureUrl: string;
  reason: string;
  statusMessage: string;
  statusType: keyof typeof ICONS.activity;
};

type EquipmentItem = {
  id: string;
  name: string;
  desc: string;
  equipmentType: string;
  pictureUrl: string;
};

export default function BorrowSummaryPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();

  const [borrow, setBorrow] = useState<BorrowDetail | null>(null);
  const [equipment, setEquipment] = useState<EquipmentItem[]>([]);
  const [cancelling, setCancelling] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    async function load() {
      try {
        const res = await fetch(`${API_BASE_URL}/borrow/${params.id}`, {
          credentials: "include",
        });
        const data = await res.json();
        if (data.success) {
          setBorrow(data.borrow);
          setEquipment(data.equipment);
        } else {
          setError(data.message || "Could not load this request.");
        }
      } catch {
        setError("Could not reach the server.");
      }
    }
    load();
  }, [params.id]);

  async function handleCancel() {
    if (!confirm("Cancel this borrow request?")) return;

    setCancelling(true);
    try {
      const res = await fetch(`${API_BASE_URL}/borrow/${params.id}/cancel`, {
        method: "POST",
        credentials: "include",
      });
      const data = await res.json();
      if (!data.success) {
        setError(data.message || "Could not cancel this request.");
        return;
      }
      router.push("/profile");
    } catch {
      setError("Could not reach the server.");
    } finally {
      setCancelling(false);
    }
  }

  if (error) {
    return <div className="bg-white rounded-2xl p-6 text-red-500">{error}</div>;
  }
  if (!borrow) {
    return <div className="text-white p-8">Loading...</div>;
  }

  const statusIcon = ICONS.activity[borrow.statusType];

  return (
    <div className="bg-white rounded-2xl p-6 flex flex-col h-full">
      <div className="flex items-baseline justify-between">
        <h1 className="text-2xl font-bold">{borrow.title}</h1>
      </div>

      <div className="flex-1 overflow-y-auto mt-4 space-y-6">
        <div>
          <h2 className="font-bold text-lg mb-2">Request Summary</h2>
          <div className="text-slate-700 space-y-1 pl-1">
            <p>Created at: {borrow.bookedAt}</p>
            <p>Scheduled at: {borrow.scheduledAt}</p>
            <p>Last updated: {borrow.statusMessageUpdateTime}</p>
            <p>Finished at: {borrow.finishedAt ?? "—"}</p>
            <p>Laboratory: {borrow.labName}</p>
          </div>
        </div>

        <div>
          <p className="text-slate-700">Reason: {borrow.reason}</p>
        </div>

        <div>
          <p className="font-semibold text-slate-900 mb-2">Status:</p>
          <div className="flex items-center gap-3">
            {statusIcon && (
              <Image src={statusIcon} alt="" width={36} height={36} />
            )}
            <span className="text-slate-700">{borrow.statusMessage}</span>
          </div>
        </div>

        <div>
          <h2 className="font-bold text-lg mb-3">Equipment List:</h2>
          <div className="divide-y divide-slate-100">
            {equipment.map((item) => (
              <div key={item.id} className="flex items-start gap-4 py-3">
                {item.pictureUrl && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={item.pictureUrl}
                    alt={item.name}
                    className="w-20 h-20 rounded-xl object-cover flex-shrink-0 bg-slate-100"
                  />
                )}
                <div className="flex-1 min-w-0">
                  <p className="font-semibold text-slate-900">{item.name}</p>
                  <p className="text-slate-500 text-sm">{item.equipmentType}</p>
                  <p className="text-slate-400 text-sm">{item.desc}</p>
                </div>
                {borrow.labPictureUrl && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={borrow.labPictureUrl}
                    alt={borrow.labName}
                    className="w-10 h-10 rounded-full object-cover flex-shrink-0 bg-slate-100"
                  />
                )}
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="flex justify-end gap-3 mt-4 flex-shrink-0">
        <button
          onClick={() => router.push("/profile")}
          aria-label="Back"
          className="w-14 h-14 rounded-full bg-slate-100 hover:bg-slate-200 transition flex items-center justify-center text-xl"
        >
          ←
        </button>

        {borrow.statusType === "pending" && (
          <button
            onClick={handleCancel}
            disabled={cancelling}
            className="px-8 h-14 rounded-full border-2 border-red-400 text-red-500 font-semibold hover:bg-red-50 transition disabled:opacity-50"
          >
            {cancelling ? "Cancelling..." : "Cancel"}
          </button>
        )}
      </div>
    </div>
  );
}