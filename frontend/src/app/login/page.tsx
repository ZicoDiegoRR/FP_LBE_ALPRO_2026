'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { KeyRound, User, Lock, AlertCircle } from 'lucide-react';
import { apiRequest, setStoredToken } from '@/lib/api';

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      const res: any = await apiRequest('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      });

      // Simpan Token JWT ke LocalStorage
      setStoredToken(res.token);

      // Simpan data User
      if (typeof window !== 'undefined') {
        localStorage.setItem('pmo_user', JSON.stringify(res.user));
      }

      // Alihkan ke Dashboard utama
      router.push('/dashboard');
    } catch (err: any) {
      setError(err.message || 'Username atau password salah');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen grid grid-cols-1 lg:grid-cols-2 bg-slate-50">
      
      {/* BAGIAN KIRI: Foto / Desain Statis */}
      <div className="hidden lg:flex relative bg-emerald-700 items-center justify-center overflow-hidden">
        {/* Ganti src di bawah ini dengan path foto desain statis Anda di folder public */}
        <img
          src="/images/login-banner.png"
          alt="Desain Sistem PMO"
          className="w-full h-full object-cover object-center"
        />
        {/* Overlay halus agar menyatu dengan estetika antarmuka */}
        <div className="absolute inset-0 bg-emerald-950/10 pointer-events-none" />
      </div>

      {/* BAGIAN KANAN: Menu & Form Login */}
      <div className="flex flex-col justify-center items-center px-6 py-12 sm:px-12 lg:px-16 bg-slate-50">
        <div className="w-full max-w-md">
          
          {/* Header Icon & Judul */}
          <div className="flex flex-col items-center text-center mb-8">
            <div className="p-3.5 bg-emerald-600/10 text-emerald-600 rounded-2xl mb-4 border border-emerald-600/20 shadow-sm">
              <KeyRound className="w-9 h-9" />
            </div>
            <h2 className="text-3xl font-bold text-slate-900 tracking-tight">
              Login Sistem PMO
            </h2>
            <p className="text-sm text-slate-500 mt-2">
              Masuk untuk mengakses layanan peminjaman terenkripsi
            </p>
          </div>

          {/* Notifikasi Error */}
          {error && (
            <div className="mb-6 p-3.5 bg-red-50 border border-red-200 rounded-xl flex items-center gap-3 text-red-600 text-sm shadow-sm">
              <AlertCircle className="w-5 h-5 flex-shrink-0" />
              <span>{error}</span>
            </div>
          )}

          {/* Form Login */}
          <form onSubmit={handleLogin} className="space-y-5">
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
                  placeholder="Username Anda"
                  className="w-full pl-11 pr-4 py-3 bg-white border border-slate-300 rounded-xl text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-600 focus:border-transparent text-sm transition-all shadow-sm"
                />
              </div>
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
                  placeholder="Password Anda"
                  className="w-full pl-11 pr-4 py-3 bg-white border border-slate-300 rounded-xl text-slate-900 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-emerald-600 focus:border-transparent text-sm transition-all shadow-sm"
                />
              </div>
            </div>

            <div className="pt-2">
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 px-4 bg-emerald-600 hover:bg-emerald-700 disabled:bg-emerald-400 disabled:cursor-not-allowed text-white font-semibold rounded-xl text-sm transition-all shadow-md shadow-emerald-600/20 flex items-center justify-center gap-2"
              >
                {loading ? 'Memverifikasi...' : 'Masuk ke Sistem'}
              </button>
            </div>
          </form>

          <p className="mt-8 text-center text-sm text-slate-600">
            Belum punya akun?{' '}
            <Link href="/register" className="text-emerald-600 hover:text-emerald-700 font-semibold transition-colors">
              Daftar di sini
            </Link>
          </p>
        </div>
      </div>

    </div>
  );
}