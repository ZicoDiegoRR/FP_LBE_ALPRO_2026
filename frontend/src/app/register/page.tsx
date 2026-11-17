'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ShieldCheck, User, Lock, AlertCircle, CheckCircle2 } from 'lucide-react';
import { apiRequest } from '@/lib/api';

export default function RegisterPage() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('user');
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);

    try {
      await apiRequest('/auth/register', {
        method: 'POST',
        body: JSON.stringify({
          username,
          password,
          role,
        }),
      });

      setSuccess('Registrasi berhasil! Mengalihkan ke halaman login...');
      setTimeout(() => {
        router.push('/login');
      }, 1500);
    } catch (err: any) {
      setError(err.message || 'Gagal melakukan pendaftaran');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen grid grid-cols-1 lg:grid-cols-2 bg-slate-50">
      
      {/* BAGIAN KIRI: Foto / Desain Statis */}
      <div className="hidden lg:flex relative bg-blue-700 items-center justify-center overflow-hidden">
        {/* Ganti src di bawah ini dengan path foto desain statis Anda di folder public */}
        <img
          src="/images/register-banner.png"
          alt="Desain Sistem PMO"
          className="w-full h-full object-cover object-center"
        />
        {/* Overlay halus agar menyatu dengan estetika antarmuka */}
        <div className="absolute inset-0 bg-blue-950/10 pointer-events-none" />
      </div>

      {/* BAGIAN KANAN: Menu & Form Pendaftaran */}
      <div className="flex flex-col justify-center items-center px-6 py-12 sm:px-12 lg:px-16 bg-slate-50">
        <div className="w-full max-w-md">
          
          {/* Header Icon & Judul */}
          <div className="flex flex-col items-center text-center mb-8">
            <div className="p-3.5 bg-blue-600/10 text-blue-600 rounded-2xl mb-4 border border-blue-600/20 shadow-sm">
              <ShieldCheck className="w-9 h-9" />
            </div>
            <h2 className="text-3xl font-bold text-slate-900 tracking-tight">
              Daftar Akun PMO
            </h2>
            <p className="text-sm text-slate-500 mt-2">
              Peminjaman Menjadi Online berbasis Enkripsi Kriptografis
            </p>
          </div>

          {/* Notifikasi Error */}
          {error && (
            <div className="mb-6 p-3.5 bg-red-50 border border-red-200 rounded-xl flex items-center gap-3 text-red-600 text-sm shadow-sm">
              <AlertCircle className="w-5 h-5 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Notifikasi Sukses */}
          {success && (
            <div className="mb-6 p-3.5 bg-emerald-50 border border-emerald-200 rounded-xl flex items-center gap-3 text-emerald-700 text-sm shadow-sm">
              <CheckCircle2 className="w-5 h-5 flex-shrink-0" />
              <span>{success}</span>
            </div>
          )}

          {/* Form Pendaftaran */}
          <form onSubmit={handleRegister} className="space-y-5">
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-2">
                Username
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 pl-3.5 flex items-center text-slate-400">
                  <User className="w-5 h-5" />
                </span>
                <input
                  type="text"
                  required
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="misal: mahasiswa1"
                  className="w-full pl-11 pr-4 py-3 bg-white border border-slate-300 rounded-xl text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent text-sm transition-all shadow-sm"
                />
              </div>
              <span className="text-[11px] text-slate-500 mt-1.5 block">
                Minimal 4 karakter alfanumerik
              </span>
            </div>

            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-2">
                Password
              </label>
              <div className="relative">
                <span className="absolute inset-y-0 left-0 pl-3.5 flex items-center text-slate-400">
                  <Lock className="w-5 h-5" />
                </span>
                <input
                  type="password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Minimal 8 karakter kuat"
                  className="w-full pl-11 pr-4 py-3 bg-white border border-slate-300 rounded-xl text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent text-sm transition-all shadow-sm"
                />
              </div>
              <span className="text-[11px] text-slate-500 mt-1.5 block">
                Wajib: Huruf besar, kecil, angka, dan simbol khusus (NIST SP 800-63B)
              </span>
            </div>

            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-2">
                Role Akun
              </label>
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="w-full px-4 py-3 bg-white border border-slate-300 rounded-xl text-slate-900 text-sm focus:outline-none focus:ring-2 focus:ring-blue-600 focus:border-transparent transition-all shadow-sm"
              >
                <option value="user">Peminjam</option>
                <option value="admin">Penyedia</option>
              </select>
            </div>

            <div className="pt-2">
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 px-4 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed text-white font-semibold rounded-xl text-sm transition-all shadow-md shadow-blue-600/20 flex items-center justify-center gap-2"
              >
                {loading ? 'Memvalidasi...' : 'Daftar Sekarang'}
              </button>
            </div>
          </form>

          <p className="mt-8 text-center text-sm text-slate-600">
            Sudah memiliki akun?{' '}
            <Link href="/login" className="text-blue-600 hover:text-blue-700 font-semibold transition-colors">
              Masuk di sini
            </Link>
          </p>
        </div>
      </div>

    </div>
  );
}