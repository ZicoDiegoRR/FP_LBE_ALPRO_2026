"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { API_BASE_URL } from "@/lib/api";

type Lab = { id: string; name: string };

type LabInfo = { id: string; name: string; pictureUrl: string };

type EquipmentItem = {
  id: string;
  name: string;
  desc: string;
  equipmentType: string;
  pictureUrl: string;
  statusId: string;
  statusName: string;
  statusSymbol: string;
};

const AVAILABLE_STATUS_ID = "JK7QYOKLJO6CEOFG4UVV";

type RequestFormData = {
  title: string;
  labId: string;
  reason: string;
  scheduledAt: string; // datetime-local value
  days: number;
  hours: number;
  minutes: number;
  seconds: number;
};

const emptyForm: RequestFormData = {
  title: "",
  labId: "",
  reason: "",
  scheduledAt: "",
  days: 0,
  hours: 0,
  minutes: 0,
  seconds: 0,
};

const boxClass =
  "w-full h-full rounded-2xl bg-slate-50 p-4 text-slate-800 placeholder-slate-400 outline-none focus:ring-2 focus:ring-slate-300 resize-none";

function NumberSpinner({
  label,
  value,
  onChange,
  max,
}: {
  label: string;
  value: number;
  onChange: (v: number) => void;
  max: number;
}) {
  function clamp(v: number) {
    if (v < 0) return 0;
    if (v > max) return max;
    return v;
  }

  return (
    <div className="flex flex-col items-center gap-1">
      <input
        type="number"
        min={0}
        max={max}
        value={value}
        onChange={(e) => onChange(clamp(Number(e.target.value)))}
        onWheel={(e) => {
          e.preventDefault();
          onChange(clamp(value + (e.deltaY < 0 ? 1 : -1)));
        }}
        className="w-16 text-center rounded-xl bg-white border border-slate-200 py-2 outline-none focus:ring-2 focus:ring-slate-300"
      />
      <span className="text-xs text-slate-500">{label}</span>
    </div>
  );
}

export default function BorrowPage() {
  const router = useRouter();
  const [step, setStep] = useState<1 | 2>(1);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState<RequestFormData>(emptyForm);
  const [agreed, setAgreed] = useState(false);
  const [labs, setLabs] = useState<Lab[]>([]);
  const [error, setError] = useState("");

  const [labInfo, setLabInfo] = useState<LabInfo | null>(null);
  const [equipment, setEquipment] = useState<EquipmentItem[]>([]);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [typeFilter, setTypeFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState("");

  useEffect(() => {
    async function loadEquipment() {
      if (step !== 2 || !form.labId) return;
      try {
        const res = await fetch(`${API_BASE_URL}/labs/${form.labId}/equipment`, {
          credentials: "include",
        });
        const data = await res.json();
        if (data.success) {
          setLabInfo(data.lab);
          setEquipment(data.equipment);
        } else {
          console.error("Failed to load equipment:", data.message);
        }
      } catch (err) {
        console.error("Could not reach /labs/:id/equipment:", err);
      }
    }
    loadEquipment();
  }, [step, form.labId]);

  const equipmentTypes = Array.from(new Set(equipment.map((e) => e.equipmentType))).filter(Boolean);
  const equipmentStatuses = Array.from(new Set(equipment.map((e) => e.statusName))).filter(Boolean);

  const filteredEquipment = equipment.filter(
    (e) =>
      (!typeFilter || e.equipmentType === typeFilter) &&
      (!statusFilter || e.statusName === statusFilter)
  );

  function toggleSelected(id: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  async function handleDone() {
    if (selected.size === 0) {
      setError("Select at least one piece of equipment.");
      return;
    }

    setError("");
    setSubmitting(true);
    try {
      const res = await fetch(`${API_BASE_URL}/borrow`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          title: form.title,
          labId: form.labId,
          reason: form.reason,
          scheduledAt: form.scheduledAt,
          days: form.days,
          hours: form.hours,
          minutes: form.minutes,
          seconds: form.seconds,
          equipmentIds: Array.from(selected),
        }),
      });
      const data = await res.json();

      if (!data.success) {
        setError(data.message || "Could not submit borrow request.");
        return;
      }

      router.push("/dashboard");
    } catch (err) {
      console.error("Could not reach /borrow:", err);
      setError("Could not reach the server.");
    } finally {
      setSubmitting(false);
    }
  }

  useEffect(() => {
    async function loadLabs() {
      try {
        const res = await fetch(`${API_BASE_URL}/labs`, {
          credentials: "include",
        });
        const data = await res.json();
        if (data.success) {
          setLabs(data.labs);
        } else {
          console.error("Failed to load labs:", data.message);
        }
      } catch (err) {
        console.error("Could not reach /labs:", err);
      }
    }
    loadLabs();
  }, []);

  function update<K extends keyof RequestFormData>(key: K, value: RequestFormData[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  function handleContinue() {
    setError("");

    if (!form.title.trim() || !form.labId || !form.reason.trim() || !form.scheduledAt) {
      setError("Please fill in every field before continuing.");
      return;
    }
    const totalSeconds =
      form.days * 86400 + form.hours * 3600 + form.minutes * 60 + form.seconds;
    if (totalSeconds <= 0) {
      setError("Borrowing duration must be more than 0.");
      return;
    }
    if (!agreed) {
      setError("Please confirm the declaration checkbox before continuing.");
      return;
    }

    // Nothing is sent to the backend / database here — this only advances
    // to the equipment-selection step, keeping everything in local state
    // until the whole request is ready to actually submit.
    setStep(2);
  }

  if (step === 2) {
    return (
      <div className="flex flex-col h-full">
        <div className="flex gap-6 flex-1 min-h-0">
          {/* Equipment list */}
          <div className="flex-1 bg-white rounded-2xl p-6 flex flex-col min-h-0">
            <div className="flex items-baseline justify-between mb-2">
              <h1 className="text-2xl font-bold">
                Select the equipment you would like to borrow:
              </h1>
            </div>

            <div className="flex-1 overflow-y-auto divide-y divide-slate-100 mt-2">
              {filteredEquipment.map((item) => (
                <label
                  key={item.id}
                  className={
                    "flex items-start gap-4 py-4 " +
                    (item.statusId === AVAILABLE_STATUS_ID
                      ? "cursor-pointer"
                      : "opacity-50 cursor-not-allowed")
                  }
                >
                  <input
                    type="checkbox"
                    checked={selected.has(item.id)}
                    disabled={item.statusId !== AVAILABLE_STATUS_ID}
                    onChange={() => toggleSelected(item.id)}
                    className="mt-2 w-4 h-4 flex-shrink-0"
                  />

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

                  <div className="flex flex-col items-end gap-1 flex-shrink-0">
                    {labInfo?.pictureUrl && (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img
                        src={labInfo.pictureUrl}
                        alt={labInfo.name}
                        className="w-10 h-10 rounded-full object-cover bg-slate-100"
                      />
                    )}
                    <span className="text-sm text-slate-500 whitespace-nowrap">
                      Status: {item.statusName} {item.statusSymbol}
                    </span>
                  </div>
                </label>
              ))}

              {filteredEquipment.length === 0 && (
                <p className="text-slate-400 py-8 text-center">
                  No equipment matches the current filters.
                </p>
              )}
            </div>
          </div>

          {/* Search / filter panel */}
          <div className="w-72 flex-shrink-0 bg-white rounded-2xl p-6">
            <h2 className="text-xl font-bold text-center">Search</h2>
            <p className="text-xs text-slate-400 text-center mt-1 mb-4">
              (Everything here is dropdown)
            </p>

            <div className="space-y-3">
              <select
                value={typeFilter}
                onChange={(e) => setTypeFilter(e.target.value)}
                className="w-full rounded-xl bg-slate-50 p-3 text-slate-700 outline-none"
              >
                <option value="">Equipment</option>
                {equipmentTypes.map((t) => (
                  <option key={t} value={t}>
                    {t}
                  </option>
                ))}
              </select>

              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="w-full rounded-xl bg-slate-50 p-3 text-slate-700 outline-none"
              >
                <option value="">Equipment status</option>
                {equipmentStatuses.map((s) => (
                  <option key={s} value={s}>
                    {s}
                  </option>
                ))}
              </select>
            </div>

            <button
              onClick={() => {
                setTypeFilter("");
                setStatusFilter("");
              }}
              className="w-full text-center text-sm text-slate-500 mt-3 hover:text-slate-700"
            >
              ✕ Clear all
            </button>
          </div>
        </div>

        <div className="flex justify-end gap-3 mt-4">
          <button
            onClick={() => setStep(1)}
            aria-label="Back"
            className="w-14 h-14 rounded-full bg-white hover:bg-slate-100 transition flex items-center justify-center text-xl"
          >
            ←
          </button>
          <button
            onClick={handleDone}
            disabled={submitting}
            className="px-8 h-14 rounded-full bg-white hover:bg-slate-100 transition font-semibold disabled:opacity-50"
          >
            {submitting ? "Submitting..." : "Done"}
          </button>
        </div>

        {error && <p className="mt-3 text-sm text-red-300">{error}</p>}
      </div>
    );
  }

  return (
    <div className="bg-white rounded-2xl p-6">
      <h1 className="text-2xl font-bold mb-6">Request Form</h1>

      <div className="grid grid-cols-2 gap-4">
        {/* Left column */}
        <div className="flex flex-col gap-4">
          <div className="h-28">
            <input
              type="text"
              placeholder="Form title"
              value={form.title}
              onChange={(e) => update("title", e.target.value)}
              className={boxClass}
            />
          </div>

          <div className="h-14">
            <select
              value={form.labId}
              onChange={(e) => update("labId", e.target.value)}
              className={boxClass + " appearance-none"}
            >
              <option value="" disabled>
                Laboratory
              </option>
              {labs.map((lab) => (
                <option key={lab.id} value={lab.id}>
                  {lab.name}
                </option>
              ))}
            </select>
          </div>

          <div className="flex-1 min-h-[160px]">
            <textarea
              placeholder="Borrowing reason"
              value={form.reason}
              onChange={(e) => update("reason", e.target.value)}
              className={boxClass}
            />
          </div>
        </div>

        {/* Right column */}
        <div className="flex flex-col gap-4">
          <div className="h-52 rounded-2xl bg-slate-50 p-4 flex flex-col justify-center">
            <label className="text-xs text-slate-500 mb-2">
              Borrowing schedule
            </label>
            <input
              type="datetime-local"
              value={form.scheduledAt}
              onChange={(e) => update("scheduledAt", e.target.value)}
              className="rounded-xl bg-white border border-slate-200 px-3 py-2 outline-none focus:ring-2 focus:ring-slate-300"
            />
          </div>

          <div className="flex-1 min-h-[160px] rounded-2xl bg-slate-50 p-4 flex flex-col">
            <label className="text-xs text-slate-500 mb-3">
              Borrowing duration
            </label>
            <div className="flex gap-4 justify-center flex-1 items-center">
              <NumberSpinner
                label="Days"
                value={form.days}
                max={999}
                onChange={(v) => update("days", v)}
              />
              <NumberSpinner
                label="Hours"
                value={form.hours}
                max={23}
                onChange={(v) => update("hours", v)}
              />
              <NumberSpinner
                label="Minutes"
                value={form.minutes}
                max={59}
                onChange={(v) => update("minutes", v)}
              />
              <NumberSpinner
                label="Seconds"
                value={form.seconds}
                max={59}
                onChange={(v) => update("seconds", v)}
              />
            </div>
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between mt-6">
        <label className="flex items-start gap-2 text-sm text-slate-700 max-w-xl">
          <input
            type="checkbox"
            checked={agreed}
            onChange={(e) => setAgreed(e.target.checked)}
            className="mt-1 w-4 h-4 flex-shrink-0"
          />
          I hereby declare that all information provided in this form is
          accurate, complete, and true to the best of my knowledge and belief.
        </label>

        <button
          onClick={handleContinue}
          aria-label="Continue"
          className="w-14 h-14 flex-shrink-0 rounded-full bg-slate-100 hover:bg-slate-200 transition flex items-center justify-center text-xl"
        >
          →
        </button>
      </div>

      {error && <p className="mt-4 text-sm text-red-500">{error}</p>}
    </div>
  );
}