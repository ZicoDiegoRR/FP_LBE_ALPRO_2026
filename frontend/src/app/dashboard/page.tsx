'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Shield, 
  Plus, 
  Users, 
  Package, 
  Key, 
  LogOut, 
  CheckCircle, 
  XCircle, 
  Clock, 
  AlertCircle,
  Lock,
  Download,
  Building,
  UserCheck
} from 'lucide-react';
import { apiRequest, clearStoredToken, getStoredToken } from '@/lib/api';

export default function DashboardPage() {
  const router = useRouter();

  // State Utama Pengguna
  const [user, setUser] = useState<{ id: number; username: string; role: string } | null>(null);
  const [loading, setLoading] = useState(true);

  // State Khusus ADMIN
  const [myOrg, setMyOrg] = useState<any>(null);
  const [pendingMembers, setPendingMembers] = useState<any[]>([]);
  const [equipments, setEquipments] = useState<any[]>([]);
  const [newOrgName, setNewOrgName] = useState('');
  const [newOrgDesc, setNewOrgDesc] = useState('');
  const [newEqName, setNewEqName] = useState('');
  const [newEqDesc, setNewEqDesc] = useState('');

  // State Khusus PEMINJAM (USER)
  const [membership, setMembership] = useState<any>(null);
  const [redeemCode, setRedeemCode] = useState('');
  const [fullname, setFullname] = useState('');
  const [identityNumber, setIdentityNumber] = useState('');
  const [phoneNumber, setPhoneNumber] = useState('');
  const [joinReason, setJoinReason] = useState('');

  // State Form Peminjaman & Kriptografi
  const [selectedEqId, setSelectedEqId] = useState('');
  const [loanPurpose, setLoanPurpose] = useState('');
  const [cryptoAlgo, setCryptoAlgo] = useState('AES');
  const [docFile, setDocFile] = useState<File | null>(null);
  const [myLoans, setMyLoans] = useState<any[]>([]);

  // Notifikasi Feedback
  const [msg, setMsg] = useState<{ type: 'error' | 'success'; text: string } | null>(null);
  const [actionLoading, setActionLoading] = useState(false);

  useEffect(() => {
    const token = getStoredToken();
    if (!token) {
      router.push('/login');
      return;
    }
    loadDashboardData();
  }, [router]);

  const loadDashboardData = async () => {
    try {
      const profileRes: any = await apiRequest('/auth/me');
      const currentUser = profileRes.user;
      setUser(currentUser);

      if (currentUser.role === 'admin') {
        // Ambil Data Organisasi Milik Admin
        const orgRes: any = await apiRequest('/admin/org');
        if (orgRes.organization) {
          setMyOrg(orgRes.organization);
          // Ambil Calon Anggota & Inventaris
          const [membersRes, eqRes]: [any, any] = await Promise.all([
            apiRequest('/admin/org/members'),
            apiRequest('/equipments'),
          ]);
          setPendingMembers(membersRes.members || []);
          setEquipments(eqRes.equipments || []);
        }
      } else {
        // Ambil Data Keanggotaan Peminjam
        const memberRes: any = await apiRequest('/orgs/my-membership');
        setMembership(memberRes.membership);

        if (memberRes.membership?.status === 'APPROVED') {
          // Jika sudah approved, ambil daftar alat & riwayat peminjaman
          const [eqRes, loansRes]: [any, any] = await Promise.all([
            apiRequest('/equipments'),
            apiRequest('/loans/my'),
          ]);
          setEquipments(eqRes.equipments || []);
          setMyLoans(loansRes.loans || []);
        }
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    clearStoredToken();
    router.push('/login');
  };

  // --- HANDLER ADMIN ---
  const handleCreateOrg = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading(true);
    setMsg(null);
    try {
      await apiRequest('/admin/org', {
        method: 'POST',
        body: JSON.stringify({ name: newOrgName, description: newOrgDesc }),
      });
      setMsg({ type: 'success', text: 'Organisasi berhasil dibuat!' });
      loadDashboardData();
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message });
    } finally {
      setActionLoading(false);
    }
  };

  const handleVerifyMember = async (memberId: number, status: 'APPROVED' | 'REJECTED') => {
    try {
      await apiRequest(`/admin/org/members/${memberId}`, {
        method: 'PATCH',
        body: JSON.stringify({ status }),
      });
      setMsg({ type: 'success', text: `Anggota berhasil di-${status.toLowerCase()}` });
      loadDashboardData();
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message });
    }
  };

  const handleAddEquipment = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading(true);
    setMsg(null);
    try {
      await apiRequest('/admin/org/equipments', {
        method: 'POST',
        body: JSON.stringify({ name: newEqName, description: newEqDesc }),
      });
      setMsg({ type: 'success', text: 'Barang berhasil ditambahkan ke inventaris!' });
      setNewEqName('');
      setNewEqDesc('');
      loadDashboardData();
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message });
    } finally {
      setActionLoading(false);
    }
  };

  // --- HANDLER USER (PEMINJAM) ---
  const handleJoinOrg = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading(true);
    setMsg(null);
    try {
      await apiRequest('/orgs/join', {
        method: 'POST',
        body: JSON.stringify({
          redeem_code: redeemCode,
          fullname,
          identity_number: identityNumber,
          phone_number: phoneNumber,
          reason: joinReason,
        }),
      });
      setMsg({ type: 'success', text: 'Pengajuan bergabung berhasil! Menunggu persetujuan Admin.' });
      loadDashboardData();
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message });
    } finally {
      setActionLoading(false);
    }
  };

  const handleCreateLoan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!docFile) {
      setMsg({ type: 'error', text: 'Dokumen bukti/syarat wajib diunggah!' });
      return;
    }
    setActionLoading(true);
    setMsg(null);
    try {
      const formData = new FormData();
      formData.append('equipment_id', selectedEqId);
      formData.append('purpose', loanPurpose);
      formData.append('crypto_algorithm', cryptoAlgo);
      formData.append('document', docFile);

      await apiRequest('/loans', {
        method: 'POST',
        body: formData,
      });

      setMsg({ type: 'success', text: `Pengajuan berhasil! Dokumen diamankan dengan algoritma ${cryptoAlgo}.` });
      setLoanPurpose('');
      setDocFile(null);
      loadDashboardData();
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message });
    } finally {
      setActionLoading(false);
    }
  };

  const handleDownloadFile = async (loanId: number, filename: string) => {
    try {
      const token = getStoredToken();
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
      const response = await fetch(`${apiUrl}/loans/${loanId}/download`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!response.ok) throw new Error('Gagal mendekripsi berkas');

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
      alert(err.message);
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
      
      {/* Navbar Header */}
      <header className="bg-white border-b border-slate-200 sticky top-0 z-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-600/10 text-emerald-600 rounded-lg">
              <Shield className="w-6 h-6" />
            </div>
            <div>
              <h1 className="font-bold text-lg text-slate-900 leading-tight">Sistem PMO</h1>
              <span className="text-xs text-slate-500 font-medium">Platform Peminjaman Terenkripsi</span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="text-right">
              <p className="text-sm font-bold text-slate-800">{user?.username}</p>
              <span className={`text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-full font-bold ${
                user?.role === 'admin' ? 'bg-purple-100 text-purple-700' : 'bg-emerald-100 text-emerald-700'
              }`}>
                {user?.role === 'admin' ? 'Penyedia (Admin)' : 'Peminjam (User)'}
              </span>
            </div>
            <button
              onClick={handleLogout}
              className="p-2 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
              title="Keluar"
            >
              <LogOut className="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      {/* Main Container */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        
        {/* Notifikasi Global */}
        {msg && (
          <div className={`p-4 rounded-xl flex items-center gap-3 text-sm shadow-sm ${
            msg.type === 'success' 
              ? 'bg-emerald-50 text-emerald-800 border border-emerald-200' 
              : 'bg-red-50 text-red-700 border border-red-200'
          }`}>
            {msg.type === 'success' ? <CheckCircle className="w-5 h-5 flex-shrink-0" /> : <AlertCircle className="w-5 h-5 flex-shrink-0" />}
            <span>{msg.text}</span>
          </div>
        )}

        {/* ======================================================== */}
        {/* TAMPILAN 1: KHUSUS ROLE ADMINISTRATOR (PENYEDIA BARANG) */}
        {/* ======================================================== */}
        {user?.role === 'admin' && (
          <div className="space-y-8">
            
            {/* JIKA ADMIN BELUM PUNYA ORGANISASI */}
            {!myOrg ? (
              <div className="max-w-xl mx-auto bg-white border border-slate-200 rounded-2xl p-8 shadow-sm text-center">
                <Building className="w-12 h-12 mx-auto text-purple-600 mb-4" />
                <h2 className="text-xl font-bold text-slate-900 mb-2">Buat Organisasi / Lab Penyedia</h2>
                <p className="text-xs text-slate-500 mb-6">
                  Sebagai Administrator, buat unit organisasi Anda terlebih dahulu agar dapat mengelola inventaris dan memverifikasi peminjam.
                </p>
                <form onSubmit={handleCreateOrg} className="space-y-4 text-left">
                  <div>
                    <label className="block text-xs font-semibold uppercase text-slate-700 mb-1">Nama Organisasi / Lab</label>
                    <input
                      type="text"
                      required
                      value={newOrgName}
                      onChange={(e) => setNewOrgName(e.target.value)}
                      placeholder="Contoh: Lab Algoritma & Pemrograman"
                      className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-xs font-semibold uppercase text-slate-700 mb-1">Deskripsi Singkat</label>
                    <textarea
                      rows={3}
                      value={newOrgDesc}
                      onChange={(e) => setNewOrgDesc(e.target.value)}
                      placeholder="Jelaskan fasilitas dan ketentuan umum peminjaman..."
                      className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-sm"
                    />
                  </div>
                  <button
                    type="submit"
                    disabled={actionLoading}
                    className="w-full py-3 bg-purple-600 hover:bg-purple-700 text-white font-semibold rounded-xl text-sm transition-all shadow-md shadow-purple-600/20"
                  >
                    {actionLoading ? 'Memproses...' : 'Daftarkan Organisasi'}
                  </button>
                </form>
              </div>
            ) : (
              /* JIKA ADMIN SUDAH MEMILIKI ORGANISASI */
              <div className="space-y-8">
                {/* Header Banner Organisasi */}
                <div className="bg-gradient-to-r from-purple-800 to-indigo-900 rounded-2xl p-6 sm:p-8 text-white flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 shadow-lg">
                  <div>
                    <div className="inline-flex items-center gap-2 px-3 py-1 bg-white/10 rounded-full text-xs font-medium mb-3 backdrop-blur-sm">
                      <Building className="w-3.5 h-3.5" /> Organisasi Aktif
                    </div>
                    <h2 className="text-2xl font-bold tracking-tight">{myOrg.name}</h2>
                    <p className="text-xs text-purple-200 mt-1 max-w-xl">{myOrg.description || 'Tidak ada deskripsi'}</p>
                  </div>
                  <div className="bg-white/10 border border-white/20 p-4 rounded-xl text-center backdrop-blur-sm">
                    <span className="text-[11px] text-purple-200 block uppercase tracking-wider font-semibold">Kode Redeem Gabung</span>
                    <span className="text-2xl font-mono font-bold tracking-widest text-emerald-300">{myOrg.redeem_code}</span>
                    <span className="text-[10px] text-purple-300 block mt-1">Bagikan kode ini ke peminjam</span>
                  </div>
                </div>

                {/* Grid: Verifikasi Anggota & Inventaris Alat */}
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                  
                  {/* Kolom 1: Verifikasi Permohonan Anggota (2 Lapis) */}
                  <div className="bg-white border border-slate-200 rounded-2xl shadow-sm p-6">
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center gap-2">
                        <UserCheck className="w-5 h-5 text-purple-600" />
                        <h3 className="font-bold text-slate-900">Verifikasi Calon Peminjam</h3>
                      </div>
                      <span className="text-xs px-2.5 py-0.5 bg-purple-100 text-purple-700 font-bold rounded-full">
                        {pendingMembers.filter(m => m.status === 'PENDING').length} Menunggu
                      </span>
                    </div>

                    {pendingMembers.length === 0 ? (
                      <p className="text-xs text-slate-400 py-8 text-center">Belum ada pengajuan anggota baru.</p>
                    ) : (
                      <div className="space-y-3 max-h-96 overflow-y-auto pr-1">
                        {pendingMembers.map((m) => (
                          <div key={m.id} className="p-4 border border-slate-200 rounded-xl bg-slate-50 flex flex-col justify-between gap-3 text-xs">
                            <div>
                              <div className="flex justify-between items-start mb-1">
                                <span className="font-bold text-slate-900 text-sm">{m.fullname}</span>
                                <span className={`px-2 py-0.5 rounded-md text-[10px] font-bold ${
                                  m.status === 'APPROVED' ? 'bg-emerald-100 text-emerald-800' :
                                  m.status === 'REJECTED' ? 'bg-red-100 text-red-800' : 'bg-amber-100 text-amber-800'
                                }`}>
                                  {m.status}
                                </span>
                              </div>
                              <p className="text-slate-600 font-medium">NIM/ID: {m.identity_number} | HP: {m.phone_number}</p>
                              <p className="text-slate-500 mt-1 italic">&ldquo;{m.reason}&rdquo;</p>
                            </div>

                            {m.status === 'PENDING' && (
                              <div className="flex gap-2 pt-2 border-t border-slate-200">
                                <button
                                  onClick={() => handleVerifyMember(m.id, 'APPROVED')}
                                  className="flex-1 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white font-semibold rounded-lg text-xs transition-colors flex items-center justify-center gap-1"
                                >
                                  <CheckCircle className="w-3.5 h-3.5" /> Setujui
                                </button>
                                <button
                                  onClick={() => handleVerifyMember(m.id, 'REJECTED')}
                                  className="flex-1 py-1.5 bg-red-600 hover:bg-red-700 text-white font-semibold rounded-lg text-xs transition-colors flex items-center justify-center gap-1"
                                >
                                  <XCircle className="w-3.5 h-3.5" /> Tolak
                                </button>
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* Kolom 2: Inventaris Barang Organisasi */}
                  <div className="bg-white border border-slate-200 rounded-2xl shadow-sm p-6 space-y-6">
                    <div className="flex items-center gap-2">
                      <Package className="w-5 h-5 text-purple-600" />
                      <h3 className="font-bold text-slate-900">Kelola Inventaris Alat</h3>
                    </div>

                    {/* Form Tambah Alat */}
                    <form onSubmit={handleAddEquipment} className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-3 text-xs">
                      <span className="font-bold text-slate-800 block">Tambah Alat Baru</span>
                      <input
                        type="text"
                        required
                        value={newEqName}
                        onChange={(e) => setNewEqName(e.target.value)}
                        placeholder="Nama Alat (contoh: Proyektor Epson, Osiloskop)"
                        className="w-full px-3 py-2 bg-white border border-slate-300 rounded-lg text-xs"
                      />
                      <input
                        type="text"
                        value={newEqDesc}
                        onChange={(e) => setNewEqDesc(e.target.value)}
                        placeholder="Deskripsi / No Seri"
                        className="w-full px-3 py-2 bg-white border border-slate-300 rounded-lg text-xs"
                      />
                      <button
                        type="submit"
                        disabled={actionLoading}
                        className="w-full py-2 bg-purple-600 hover:bg-purple-700 text-white font-semibold rounded-lg text-xs transition-colors flex items-center justify-center gap-1"
                      >
                        <Plus className="w-3.5 h-3.5" /> Tambah ke Inventaris
                      </button>
                    </form>

                    {/* Daftar Alat Saat Ini */}
                    <div className="space-y-2">
                      <span className="text-xs font-bold text-slate-700 block">Daftar Alat ({equipments.length})</span>
                      {equipments.length === 0 ? (
                        <p className="text-xs text-slate-400 py-3 text-center">Belum ada alat yang didaftarkan.</p>
                      ) : (
                        <div className="divide-y divide-slate-100 max-h-56 overflow-y-auto">
                          {equipments.map((eq) => (
                            <div key={eq.id} className="py-2.5 flex justify-between items-center text-xs">
                              <div>
                                <span className="font-semibold text-slate-900 block">{eq.name}</span>
                                <span className="text-slate-400 text-[11px]">{eq.description || 'Tanpa deskripsi'}</span>
                              </div>
                              <span className="px-2 py-0.5 bg-emerald-50 text-emerald-700 font-semibold rounded text-[10px] border border-emerald-200">
                                {eq.status}
                              </span>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>

                </div>
              </div>
            )}

          </div>
        )}

        {/* ======================================================== */}
        {/* TAMPILAN 2: KHUSUS ROLE PEMINJAM (USER / MAHASISWA)     */}
        {/* ======================================================== */}
        {user?.role === 'user' && (
          <div className="space-y-8">
            
            {/* FASE 1: BELUM BERGABUNG DENGAN ORGANISASI */}
            {!membership && (
              <div className="max-w-xl mx-auto bg-white border border-slate-200 rounded-2xl p-8 shadow-sm">
                <div className="text-center mb-6">
                  <div className="p-3 bg-emerald-600/10 text-emerald-600 rounded-2xl inline-block mb-3">
                    <Key className="w-8 h-8" />
                  </div>
                  <h2 className="text-xl font-bold text-slate-900">Gabung ke Organisasi / Lab</h2>
                  <p className="text-xs text-slate-500 mt-1">
                    Masukkan Kode Redeem dan lengkapi biodata verifikasi 2-lapis Anda untuk dapat meminjam alat.
                  </p>
                </div>

                <form onSubmit={handleJoinOrg} className="space-y-4 text-xs">
                  <div>
                    <label className="block font-semibold uppercase text-slate-700 mb-1">Kode Redeem Organisasi</label>
                    <input
                      type="text"
                      required
                      value={redeemCode}
                      onChange={(e) => setRedeemCode(e.target.value.toUpperCase())}
                      placeholder="Contoh: 8 Karakter Kode dari Admin Lab"
                      className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl font-mono uppercase font-bold text-sm tracking-widest text-emerald-700"
                    />
                  </div>
                  <div>
                    <label className="block font-semibold uppercase text-slate-700 mb-1">Nama Lengkap</label>
                    <input
                      type="text"
                      required
                      value={fullname}
                      onChange={(e) => setFullname(e.target.value)}
                      placeholder="Sesuai kartu identitas mahasiswa"
                      className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl"
                    />
                  </div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="block font-semibold uppercase text-slate-700 mb-1">NIM / NIK</label>
                      <input
                        type="text"
                        required
                        value={identityNumber}
                        onChange={(e) => setIdentityNumber(e.target.value)}
                        placeholder="Contoh: 5025211..."
                        className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl"
                      />
                    </div>
                    <div>
                      <label className="block font-semibold uppercase text-slate-700 mb-1">No WhatsApp / HP</label>
                      <input
                        type="tel"
                        required
                        value={phoneNumber}
                        onChange={(e) => setPhoneNumber(e.target.value)}
                        placeholder="08123456789"
                        className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl"
                      />
                    </div>
                  </div>
                  <div>
                    <label className="block font-semibold uppercase text-slate-700 mb-1">Alasan Bergabung / Keperluan Umum</label>
                    <textarea
                      rows={2}
                      required
                      value={joinReason}
                      onChange={(e) => setJoinReason(e.target.value)}
                      placeholder="Untuk praktikum Alpro semester ini..."
                      className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl"
                    />
                  </div>
                  <button
                    type="submit"
                    disabled={actionLoading}
                    className="w-full py-3 bg-emerald-600 hover:bg-emerald-700 text-white font-semibold rounded-xl text-sm transition-all shadow-md shadow-emerald-600/20"
                  >
                    {actionLoading ? 'Mengirim Data...' : 'Kirim Biodata & Ajukan Verifikasi'}
                  </button>
                </form>
              </div>
            )}

            {/* FASE 2: SUDAH DAFTAR TAPI MASIH STATUS PENDING */}
            {membership && membership.status === 'PENDING' && (
              <div className="max-w-xl mx-auto bg-amber-50 border border-amber-200 rounded-2xl p-8 text-center text-amber-900 shadow-sm">
                <Clock className="w-12 h-12 mx-auto text-amber-600 mb-3 animate-pulse" />
                <h2 className="text-xl font-bold mb-2">Menunggu Persetujuan Admin Lab</h2>
                <p className="text-xs text-amber-700 leading-relaxed mb-4">
                  Anda telah mengajukan bergabung ke organisasi <span className="font-bold">{membership.organization?.name}</span>. 
                  Admin sedang memverifikasi biodata dan identitas Anda. Harap tunggu hingga akun Anda disetujui untuk membuka peminjaman alat.
                </p>
                <div className="inline-block px-4 py-1.5 bg-amber-200/60 rounded-full text-xs font-bold text-amber-800">
                  Status: PENDING VERIFICATION
                </div>
              </div>
            )}

            {/* FASE 3: SUDAH APPROVED -> BISA MEMINJAM BARANG & FILE TERENKRIPSI */}
            {membership && membership.status === 'APPROVED' && (
              <div className="space-y-8">
                {/* Banner Status Anggota */}
                <div className="bg-emerald-50 border border-emerald-200 rounded-2xl p-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                  <div>
                    <span className="text-[10px] uppercase tracking-wider font-bold text-emerald-700 bg-emerald-100 px-2 py-0.5 rounded-full">
                      Anggota Terverifikasi
                    </span>
                    <h2 className="text-xl font-bold text-slate-900 mt-1">{membership.organization?.name}</h2>
                    <p className="text-xs text-slate-500">Peminjam: {membership.fullname} ({membership.identity_number})</p>
                  </div>
                  <span className="px-3 py-1 bg-emerald-600 text-white text-xs font-semibold rounded-xl flex items-center gap-1.5">
                    <CheckCircle className="w-4 h-4" /> Akses Peminjaman Aktif
                  </span>
                </div>

                {/* Form Peminjaman Alat & Kriptografi */}
                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                  <div className="lg:col-span-2 bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
                    <h3 className="font-bold text-slate-900 text-base mb-4 flex items-center gap-2">
                      <Lock className="w-5 h-5 text-emerald-600" /> Form Pengajuan Peminjaman Alat Terenkripsi
                    </h3>

                    {equipments.length === 0 ? (
                      <p className="text-xs text-slate-400 py-6 text-center">Organisasi ini belum memiliki barang yang siap dipinjam.</p>
                    ) : (
                      <form onSubmit={handleCreateLoan} className="space-y-4 text-xs">
                        <div>
                          <label className="block font-semibold uppercase text-slate-700 mb-1">Pilih Alat / Fasilitas</label>
                          <select
                            required
                            value={selectedEqId}
                            onChange={(e) => setSelectedEqId(e.target.value)}
                            className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-slate-900 font-medium"
                          >
                            <option value="">-- Pilih Barang --</option>
                            {equipments.map((eq) => (
                              <option key={eq.id} value={eq.id}>
                                {eq.name} ({eq.status})
                              </option>
                            ))}
                          </select>
                        </div>

                        <div>
                          <label className="block font-semibold uppercase text-slate-700 mb-1">Keperluan Peminjaman</label>
                          <textarea
                            rows={2}
                            required
                            value={loanPurpose}
                            onChange={(e) => setLoanPurpose(e.target.value)}
                            placeholder="Jelaskan tujuan pemakaian alat dan durasinya..."
                            className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl"
                          />
                        </div>

                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                          <div>
                            <label className="block font-semibold uppercase text-slate-700 mb-1">Algoritma Enkripsi Berkas</label>
                            <select
                              value={cryptoAlgo}
                              onChange={(e) => setCryptoAlgo(e.target.value)}
                              className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl font-bold text-emerald-700"
                            >
                              <option value="AES">AES-256-GCM (Modern AEAD)</option>
                              <option value="DES">DES-CBC (Legacy 56-bit)</option>
                              <option value="RC4">RC4 (Stream Cipher)</option>
                            </select>
                          </div>
                          <div>
                            <label className="block font-semibold uppercase text-slate-700 mb-1">Berkas Syarat (KTM/Surat)</label>
                            <input
                              type="file"
                              required
                              onChange={(e) => setDocFile(e.target.files?.[0] || null)}
                              className="w-full text-slate-500 file:mr-3 file:py-1.5 file:px-3 file:rounded-xl file:border-0 file:bg-emerald-50 file:text-emerald-700 file:font-semibold"
                            />
                          </div>
                        </div>

                        <button
                          type="submit"
                          disabled={actionLoading}
                          className="w-full py-3 bg-emerald-600 hover:bg-emerald-700 text-white font-semibold rounded-xl text-sm transition-all shadow-md shadow-emerald-600/20"
                        >
                          {actionLoading ? 'Mengenkripsi & Mengunggah...' : 'Enkripsi & Ajukan Peminjaman'}
                        </button>
                      </form>
                    )}
                  </div>

                  {/* Ringkasan Edukasi Kriptografi */}
                  <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm text-xs space-y-3">
                    <h4 className="font-bold text-slate-900 text-sm">Standar Kriptografi PMO</h4>
                    <p className="text-slate-500 leading-relaxed">
                      Seluruh berkas persyaratan yang Anda unggah otomatis dienkripsi sebelum disimpan di server fisik.
                    </p>
                    <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl">
                      <span className="font-bold text-emerald-700 block">AES-256-GCM</span>
                      <p className="text-slate-500 mt-0.5">Enkripsi terautentikasi (AEAD) anti manipulasi file.</p>
                    </div>
                    <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl">
                      <span className="font-bold text-amber-700 block">DES-CBC</span>
                      <p className="text-slate-500 mt-0.5">Block cipher klasik kelipatan 8 byte dengan padding PKCS#7.</p>
                    </div>
                    <div className="p-3 bg-slate-50 border border-slate-200 rounded-xl">
                      <span className="font-bold text-blue-700 block">RC4 Stream</span>
                      <p className="text-slate-500 mt-0.5">Stream cipher berbasis operasi XOR langsung pada byte data.</p>
                    </div>
                  </div>
                </div>

                {/* Tabel Riwayat Peminjaman User */}
                <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
                  <div className="px-6 py-4 border-b border-slate-200 font-bold text-sm text-slate-900">
                    Riwayat Peminjaman Anda
                  </div>
                  {myLoans.length === 0 ? (
                    <p className="text-xs text-slate-400 py-8 text-center">Belum ada riwayat peminjaman alat.</p>
                  ) : (
                    <table className="w-full text-left text-xs">
                      <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider font-semibold border-b border-slate-200">
                        <tr>
                          <th className="py-3 px-6">Barang</th>
                          <th className="py-3 px-6">Keperluan</th>
                          <th className="py-3 px-6">Algoritma</th>
                          <th className="py-3 px-6">Status</th>
                          <th className="py-3 px-6 text-right">Aksi</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100">
                        {myLoans.map((loan) => (
                          <tr key={loan.id} className="hover:bg-slate-50">
                            <td className="py-3 px-6 font-semibold">{loan.equipment?.name || '-'}</td>
                            <td className="py-3 px-6 text-slate-600">{loan.purpose}</td>
                            <td className="py-3 px-6 font-bold text-emerald-700">{loan.crypto_algorithm}</td>
                            <td className="py-3 px-6">
                              <span className="px-2 py-0.5 bg-amber-50 text-amber-700 font-bold rounded-full border border-amber-200">
                                {loan.status}
                              </span>
                            </td>
                            <td className="py-3 px-6 text-right">
                              <button
                                onClick={() => handleDownloadFile(loan.id, loan.original_filename)}
                                className="inline-flex items-center gap-1 px-3 py-1 bg-slate-100 hover:bg-emerald-50 text-slate-700 hover:text-emerald-700 rounded-lg font-semibold border border-slate-200 transition-colors"
                              >
                                <Download className="w-3.5 h-3.5" /> Dekripsi
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  )}
                </div>

              </div>
            )}

          </div>
        )}

      </main>
    </div>
  );
}