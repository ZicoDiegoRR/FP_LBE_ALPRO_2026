'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Shield, 
  Upload, 
  Download, 
  FileText, 
  LogOut, 
  CheckCircle, 
  Clock, 
  AlertCircle,
  Lock,
  Layers
} from 'lucide-react';
import { apiRequest, clearStoredToken, getStoredToken } from '@/lib/api';

interface Loan {
  id: number;
  item_name: string;
  description: string;
  status: string;
  original_filename: string;
  crypto_algorithm: string;
  created_at: string;
}

export default function DashboardPage() {
  const router = useRouter();
  
  // State User & Loans
  const [user, setUser] = useState<{ id: number; username: string; role: string } | null>(null);
  const [loans, setLoans] = useState<Loan[]>([]);
  const [loading, setLoading] = useState(true);

  // State Form Upload
  const [itemName, setItemName] = useState('');
  const [description, setDescription] = useState('');
  const [algorithm, setAlgorithm] = useState('AES');
  const [file, setFile] = useState<File | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [formMsg, setFormMsg] = useState<{ type: 'error' | 'success'; text: string } | null>(null);

  // 1. Cek Sesi Autentikasi & Ambil Data
  useEffect(() => {
    const token = getStoredToken();
    if (!token) {
      router.push('/login');
      return;
    }

    fetchUserData();
    fetchUserLoans();
  }, [router]);

  const fetchUserData = async () => {
    try {
      const res: any = await apiRequest('/auth/me');
      setUser(res.user);
    } catch {
      handleLogout();
    }
  };

  const fetchUserLoans = async () => {
    try {
      const res: any = await apiRequest('/loans/my');
      setLoans(res.loans || []);
    } catch (err: any) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    clearStoredToken();
    router.push('/login');
  };

  // 2. Submit Pengajuan Peminjaman + Upload File Terenkripsi
  const handleCreateLoan = async (e: React.FormEvent) => {
    e.preventDefault();
    setFormMsg(null);

    if (!file) {
      setFormMsg({ type: 'error', text: 'Pilih file dokumen pendukung terlebih dahulu' });
      return;
    }

    setSubmitting(true);

    try {
      const formData = new FormData();
      formData.append('item_name', itemName);
      formData.append('description', description);
      formData.append('crypto_algorithm', algorithm);
      formData.append('document', file);

      await apiRequest('/loans', {
        method: 'POST',
        body: formData,
      });

      setFormMsg({ 
        type: 'success', 
        text: `Pengajuan berhasil! File telah diamankan dengan algoritma ${algorithm} di server.` 
      });

      // Reset form & refresh tabel
      setItemName('');
      setDescription('');
      setFile(null);
      fetchUserLoans();
    } catch (err: any) {
      setFormMsg({ type: 'error', text: err.message || 'Gagal mengirim pengajuan' });
    } finally {
      setSubmitting(false);
    }
  };

  // 3. Download & Dekripsi File On-The-Fly
  const handleDownload = async (loanId: number, filename: string) => {
    try {
      const token = getStoredToken();
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

      // Panggil endpoint download biner
      const response = await fetch(`${apiUrl}/loans/${loanId}/download`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Gagal mendekripsi file atau akses ditolak');
      }

      // Konversi respon biner menjadi Blob dan picu unduhan di browser
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err: any) {
      alert(err.message || 'Terjadi kesalahan saat mendownload file');
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-emerald-600"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      
      {/* Top Navbar */}
      <header className="bg-white border-b border-slate-200 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-600/10 text-emerald-600 rounded-lg">
              <Shield className="w-6 h-6" />
            </div>
            <div>
              <h1 className="font-bold text-lg text-slate-900">Sistem PMO</h1>
              <span className="text-xs text-slate-500 font-medium">Platform Peminjaman Terenkripsi</span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="text-right hidden sm:block">
              <p className="text-sm font-semibold text-slate-800">{user?.username}</p>
              <span className="text-xs uppercase tracking-wider px-2 py-0.5 bg-emerald-100 text-emerald-800 font-bold rounded-full">
                {user?.role}
              </span>
            </div>
            <button
              onClick={handleLogout}
              className="p-2 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
              title="Keluar"
            >
              <LogOut className="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        
        {/* Grid: Form Peminjaman & Informasi Kriptografi */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          
          {/* Kolom 1-2: Form Pengajuan */}
          <div className="lg:col-span-2 bg-white border border-slate-200 rounded-2xl shadow-sm p-6 sm:p-8">
            <div className="flex items-center gap-2 text-emerald-600 font-semibold mb-6">
              <Upload className="w-5 h-5" />
              <h2 className="text-lg text-slate-900">Form Pengajuan Peminjaman Barang</h2>
            </div>

            {formMsg && (
              <div className={`mb-6 p-4 rounded-xl flex items-center gap-3 text-sm ${
                formMsg.type === 'success' 
                  ? 'bg-emerald-50 text-emerald-800 border border-emerald-200' 
                  : 'bg-red-50 text-red-700 border border-red-200'
              }`}>
                {formMsg.type === 'success' ? <CheckCircle className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
                <span>{formMsg.text}</span>
              </div>
            )}

            <form onSubmit={handleCreateLoan} className="space-y-5">
              <div>
                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-2">
                  Nama Barang / Fasilitas
                </label>
                <input
                  type="text"
                  required
                  value={itemName}
                  onChange={(e) => setItemName(e.target.value)}
                  placeholder="Contoh: Proyektor Lab, Ruang Rapat, Kamera DSLR"
                  className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-slate-900 focus:outline-none focus:ring-2 focus:ring-emerald-600 text-sm"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-2">
                  Keperluan / Deskripsi
                </label>
                <textarea
                  rows={3}
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Jelaskan tujuan peminjaman dan estimasi durasi pemakaian..."
                  className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-slate-900 focus:outline-none focus:ring-2 focus:ring-emerald-600 text-sm"
                />
              </div>

              {/* Pemilihan Algoritma Enkripsi */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-2">
                    Algoritma Enkripsi Dokumen
                  </label>
                  <select
                    value={algorithm}
                    onChange={(e) => setAlgorithm(e.target.value)}
                    className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-slate-900 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-600 font-medium"
                  >
                    <option value="AES">AES-256-GCM (Direkomendasikan)</option>
                    <option value="DES">DES-CBC (Legacy 56-bit)</option>
                    <option value="RC4">RC4 (Stream Cipher)</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-semibold uppercase tracking-wider text-slate-700 mb-2">
                    Upload Dokumen Persyaratan (KTM/Surat)
                  </label>
                  <input
                    type="file"
                    required
                    onChange={(e) => setFile(e.target.files?.[0] || null)}
                    className="w-full text-sm text-slate-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-semibold file:bg-emerald-50 file:text-emerald-700 hover:file:bg-emerald-100"
                  />
                </div>
              </div>

              <div className="pt-2">
                <button
                  type="submit"
                  disabled={submitting}
                  className="w-full py-3 bg-emerald-600 hover:bg-emerald-700 disabled:bg-emerald-400 text-white font-semibold rounded-xl text-sm transition-all shadow-md shadow-emerald-600/20 flex items-center justify-center gap-2"
                >
                  <Lock className="w-4 h-4" />
                  {submitting ? 'Mengenkripsi & Mengunggah Dokumen...' : 'Enkripsi & Ajukan Peminjaman'}
                </button>
              </div>
            </form>
          </div>

          {/* Kolom 3: Security Info Card */}
          <div className="bg-white border border-slate-200 rounded-2xl shadow-sm p-6 space-y-4">
            <div className="flex items-center gap-2 text-slate-900 font-semibold">
              <Layers className="w-5 h-5 text-emerald-600" />
              <h3 className="text-base">Mekanisme Kriptografi</h3>
            </div>
            
            <p className="text-xs text-slate-600 leading-relaxed">
              Dokumen persyaratan yang diunggah akan langsung dienkripsi di memori backend (Zero-Plaintext at Rest).
            </p>

            <div className="space-y-3 pt-2 text-xs">
              <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl">
                <span className="font-bold text-emerald-700 block mb-1">AES-256-GCM</span>
                <p className="text-slate-500">Standar enkripsi tertinggi dengan tag autentikasi integritas data.</p>
              </div>

              <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl">
                <span className="font-bold text-amber-700 block mb-1">DES-CBC</span>
                <p className="text-slate-500">Block cipher klasik dengan kunci 56-bit dan bantalan PKCS#7.</p>
              </div>

              <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl">
                <span className="font-bold text-blue-700 block mb-1">RC4 Stream</span>
                <p className="text-slate-500">Stream cipher berbasis operasi XOR kontinu tanpa penambahan padding.</p>
              </div>
            </div>
          </div>
        </div>

        {/* Tabel Riwayat Peminjaman & Dokumen Terenkripsi */}
        <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
          <div className="px-6 py-5 border-b border-slate-200 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <FileText className="w-5 h-5 text-emerald-600" />
              <h2 className="text-base font-bold text-slate-900">Riwayat Peminjaman & Berkas Terenkripsi</h2>
            </div>
            <span className="text-xs bg-slate-100 text-slate-600 px-3 py-1 rounded-full font-medium">
              Total: {loans.length} Pengajuan
            </span>
          </div>

          {loans.length === 0 ? (
            <div className="p-12 text-center text-slate-500">
              <FileText className="w-12 h-12 mx-auto text-slate-300 mb-3" />
              <p className="text-sm">Belum ada riwayat pengajuan peminjaman.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse text-sm">
                <thead>
                  <tr className="bg-slate-50 border-b border-slate-200 text-slate-500 text-xs uppercase tracking-wider font-semibold">
                    <th className="py-3.5 px-6">Barang</th>
                    <th className="py-3.5 px-6">Deskripsi</th>
                    <th className="py-3.5 px-6">Algoritma</th>
                    <th className="py-3.5 px-6">Status</th>
                    <th className="py-3.5 px-6">File Berkas</th>
                    <th className="py-3.5 px-6 text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {loans.map((loan) => (
                    <tr key={loan.id} className="hover:bg-slate-50 transition-colors">
                      <td className="py-4 px-6 font-semibold text-slate-900">{loan.item_name}</td>
                      <td className="py-4 px-6 text-slate-600 max-w-xs truncate">{loan.description || '-'}</td>
                      <td className="py-4 px-6">
                        <span className={`px-2.5 py-1 rounded-md text-xs font-bold ${
                          loan.crypto_algorithm === 'AES' ? 'bg-emerald-100 text-emerald-800' :
                          loan.crypto_algorithm === 'DES' ? 'bg-amber-100 text-amber-800' :
                          'bg-blue-100 text-blue-800'
                        }`}>
                          {loan.crypto_algorithm}
                        </span>
                      </td>
                      <td className="py-4 px-6">
                        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
                          <Clock className="w-3.5 h-3.5" />
                          {loan.status}
                        </span>
                      </td>
                      <td className="py-4 px-6 text-slate-500 text-xs">
                        <span className="truncate block max-w-[150px]" title={loan.original_filename}>
                          {loan.original_filename}
                        </span>
                      </td>
                      <td className="py-4 px-6 text-right">
                        <button
                          onClick={() => handleDownload(loan.id, loan.original_filename)}
                          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-emerald-50 text-slate-700 hover:text-emerald-700 rounded-lg text-xs font-semibold transition-colors border border-slate-200"
                        >
                          <Download className="w-3.5 h-3.5" />
                          Dekripsi & Unduh
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

      </main>
    </div>
  );
}