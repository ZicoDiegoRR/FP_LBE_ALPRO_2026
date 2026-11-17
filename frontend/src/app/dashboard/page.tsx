'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { 
  Shield, 
  Plus, 
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
  UserCheck,
  FileCheck,
  FileText,
  Fingerprint,
  Share2,
  FolderLock
} from 'lucide-react';
import { apiRequest, clearStoredToken, getStoredToken } from '@/lib/api';

export default function DashboardPage() {
  const router = useRouter();

  // State Utama Pengguna & Navigasi Tab
  const [user, setUser] = useState<{ id: number; username: string; role: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<'inventory' | 'vault'>('inventory');

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

  // State Peminjaman Alat
  const [selectedEqId, setSelectedEqId] = useState('');
  const [loanPurpose, setLoanPurpose] = useState('');
  const [cryptoAlgo, setCryptoAlgo] = useState('AES');
  const [docFile, setDocFile] = useState<File | null>(null);
  const [myLoans, setMyLoans] = useState<any[]>([]);

  // State BRANKAS DIGITAL & DIGITAL SIGNATURE (VAULT)
  const [hasRSAKey, setHasRSAKey] = useState(false);
  const [publicKeyPEM, setPublicKeyPEM] = useState('');
  const [vaultFiles, setVaultFiles] = useState<any[]>([]);
  const [vaultUploadFile, setVaultUploadFile] = useState<File | null>(null);

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

      // Ambil Berkas Brankas Pribadi
      loadVaultData();

      if (currentUser.role === 'admin') {
        const orgRes: any = await apiRequest('/admin/org');
        if (orgRes.organization) {
          setMyOrg(orgRes.organization);
          const [membersRes, eqRes]: [any, any] = await Promise.all([
            apiRequest('/admin/org/members'),
            apiRequest('/equipments'),
          ]);
          setPendingMembers(membersRes.members || []);
          setEquipments(eqRes.equipments || []);
        }
      } else {
        const memberRes: any = await apiRequest('/orgs/my-membership');
        setMembership(memberRes.membership);

        if (memberRes.membership?.status === 'APPROVED') {
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

  const loadVaultData = async () => {
    try {
      const res: any = await apiRequest('/vault/files');
      setVaultFiles(res.files || []);
    } catch (err) {
      console.error(err);
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

      await apiRequest('/loans', { method: 'POST', body: formData });
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

  // --- HANDLER VAULT & DIGITAL SIGNATURE ---
  const handleGenerateRSA = async () => {
    setActionLoading(true);
    setMsg(null);
    try {
      const res: any = await apiRequest('/vault/keys/generate', { method: 'POST' });
      setPublicKeyPEM(res.public_key);
      setHasRSAKey(true);
      setMsg({ type: 'success', text: 'Pasangan Kunci RSA 2048-bit berhasil diaktifkan!' });
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message });
    } finally {
      setActionLoading(false);
    }
  };

  const handleUploadVault = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!vaultUploadFile) return;
    setActionLoading(true);
    setMsg(null);
    try {
      const formData = new FormData();
      formData.append('document', vaultUploadFile);

      await apiRequest('/vault/files', { method: 'POST', body: formData });
      setMsg({ type: 'success', text: 'Dokumen berhasil dienkripsi dan disimpan di Brankas Digital!' });
      setVaultUploadFile(null);
      loadVaultData();
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message });
    } finally {
      setActionLoading(false);
    }
  };

  const handleSignFile = async (fileID: number) => {
    setActionLoading(true);
    setMsg(null);
    try {
      await apiRequest(`/vault/files/${fileID}/sign`, { method: 'POST' });
      setMsg({ type: 'success', text: 'Dokumen resmi ditandatangani secara digital dengan RSA Private Key!' });
      loadVaultData();
    } catch (err: any) {
      setMsg({ type: 'error', text: err.message });
    } finally {
      setActionLoading(false);
    }
  };

  const handleVerifyFile = async (fileID: number) => {
    try {
      const res: any = await apiRequest(`/vault/files/${fileID}/verify`);
      if (res.valid) {
        alert(`VERIFIKASI BERHASIL ✅\n\nStatus: ${res.message}\nPenandatangan: ${res.signer}\nHash SHA-256: ${res.file_hash}`);
      } else {
        alert(`VERIFIKASI GAGAL ❌\n\n${res.message}`);
      }
    } catch (err: any) {
      alert('Error saat verifikasi: ' + err.message);
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
              <span className="text-xs text-slate-500 font-medium">Platform Peminjaman & Kriptografi Terpadu</span>
            </div>
          </div>

          {/* Tab Navigasi Utama */}
          <div className="flex items-center bg-slate-100 p-1 rounded-xl">
            <button
              onClick={() => setActiveTab('inventory')}
              className={`px-4 py-1.5 rounded-lg text-xs font-bold transition-all ${
                activeTab === 'inventory' ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-500 hover:text-slate-900'
              }`}
            >
              Layanan Peminjaman
            </button>
            <button
              onClick={() => setActiveTab('vault')}
              className={`px-4 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center gap-1.5 ${
                activeTab === 'vault' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-500 hover:text-slate-900'
              }`}
            >
              <FolderLock className="w-3.5 h-3.5" /> Brankas Kriptografi
            </button>
          </div>

          <div className="flex items-center gap-4">
            <div className="text-right hidden sm:block">
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
        
        {/* Notifikasi Feedback Global */}
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
        {/* TAB 1: LAYANAN PEMINJAMAN BARANG & ORGANISASI           */}
        {/* ======================================================== */}
        {activeTab === 'inventory' && (
          <>
            {/* TAMPILAN ADMIN */}
            {user?.role === 'admin' && (
              <div className="space-y-8">
                {!myOrg ? (
                  <div className="max-w-xl mx-auto bg-white border border-slate-200 rounded-2xl p-8 shadow-sm text-center">
                    <Building className="w-12 h-12 mx-auto text-purple-600 mb-4" />
                    <h2 className="text-xl font-bold text-slate-900 mb-2">Buat Organisasi / Lab Penyedia</h2>
                    <p className="text-xs text-slate-500 mb-6">
                      Sebagai Administrator, buat unit organisasi Anda terlebih dahulu agar dapat mengelola inventaris.
                    </p>
                    <form onSubmit={handleCreateOrg} className="space-y-4 text-left text-xs">
                      <div>
                        <label className="block font-semibold uppercase text-slate-700 mb-1">Nama Organisasi / Lab</label>
                        <input
                          type="text"
                          required
                          value={newOrgName}
                          onChange={(e) => setNewOrgName(e.target.value)}
                          placeholder="Contoh: Lab Algoritma & Pemrograman"
                          className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl"
                        />
                      </div>
                      <div>
                        <label className="block font-semibold uppercase text-slate-700 mb-1">Deskripsi Singkat</label>
                        <textarea
                          rows={3}
                          value={newOrgDesc}
                          onChange={(e) => setNewOrgDesc(e.target.value)}
                          placeholder="Jelaskan fasilitas dan ketentuan umum peminjaman..."
                          className="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 rounded-xl"
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
                  <div className="space-y-8">
                    <div className="bg-gradient-to-r from-purple-800 to-indigo-900 rounded-2xl p-6 sm:p-8 text-white flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 shadow-lg">
                      <div>
                        <div className="inline-flex items-center gap-2 px-3 py-1 bg-white/10 rounded-full text-xs font-medium mb-3 backdrop-blur-sm">
                          <Building className="w-3.5 h-3.5" /> Organisasi Aktif
                        </div>
                        <h2 className="text-2xl font-bold tracking-tight">{myOrg.name}</h2>
                        <p className="text-xs text-purple-200 mt-1 max-w-xl">{myOrg.description || 'Tanpa deskripsi'}</p>
                      </div>
                      <div className="bg-white/10 border border-white/20 p-4 rounded-xl text-center backdrop-blur-sm">
                        <span className="text-[11px] text-purple-200 block uppercase tracking-wider font-semibold">Kode Redeem Gabung</span>
                        <span className="text-2xl font-mono font-bold tracking-widest text-emerald-300">{myOrg.redeem_code}</span>
                      </div>
                    </div>

                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                      {/* Verifikasi Anggota */}
                      <div className="bg-white border border-slate-200 rounded-2xl shadow-sm p-6">
                        <div className="flex items-center justify-between mb-4">
                          <div className="flex items-center gap-2">
                            <UserCheck className="w-5 h-5 text-purple-600" />
                            <h3 className="font-bold text-slate-900 text-sm">Verifikasi Calon Peminjam</h3>
                          </div>
                          <span className="text-xs px-2.5 py-0.5 bg-purple-100 text-purple-700 font-bold rounded-full">
                            {pendingMembers.filter(m => m.status === 'PENDING').length} Menunggu
                          </span>
                        </div>
                        {pendingMembers.length === 0 ? (
                          <p className="text-xs text-slate-400 py-8 text-center">Belum ada pengajuan anggota.</p>
                        ) : (
                          <div className="space-y-3 max-h-80 overflow-y-auto">
                            {pendingMembers.map((m) => (
                              <div key={m.id} className="p-3.5 border border-slate-200 rounded-xl bg-slate-50 text-xs">
                                <div className="flex justify-between items-start mb-1">
                                  <span className="font-bold text-slate-900">{m.fullname}</span>
                                  <span className={`px-2 py-0.5 rounded text-[10px] font-bold ${
                                    m.status === 'APPROVED' ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800'
                                  }`}>{m.status}</span>
                                </div>
                                <p className="text-slate-500">NIM: {m.identity_number} | HP: {m.phone_number}</p>
                                <p className="text-slate-600 italic mt-1">&ldquo;{m.reason}&rdquo;</p>
                                {m.status === 'PENDING' && (
                                  <div className="flex gap-2 mt-2 pt-2 border-t border-slate-200">
                                    <button onClick={() => handleVerifyMember(m.id, 'APPROVED')} className="flex-1 py-1 bg-emerald-600 text-white rounded font-bold text-[11px]">Setujui</button>
                                    <button onClick={() => handleVerifyMember(m.id, 'REJECTED')} className="flex-1 py-1 bg-red-600 text-white rounded font-bold text-[11px]">Tolak</button>
                                  </div>
                                )}
                              </div>
                            ))}
                          </div>
                        )}
                      </div>

                      {/* Inventaris Alat */}
                      <div className="bg-white border border-slate-200 rounded-2xl shadow-sm p-6 space-y-4">
                        <div className="flex items-center gap-2">
                          <Package className="w-5 h-5 text-purple-600" />
                          <h3 className="font-bold text-slate-900 text-sm">Inventaris Barang Lab</h3>
                        </div>
                        <form onSubmit={handleAddEquipment} className="p-3.5 bg-slate-50 border border-slate-200 rounded-xl space-y-2 text-xs">
                          <input type="text" required value={newEqName} onChange={(e) => setNewEqName(e.target.value)} placeholder="Nama Alat" className="w-full px-3 py-1.5 bg-white border border-slate-300 rounded" />
                          <input type="text" value={newEqDesc} onChange={(e) => setNewEqDesc(e.target.value)} placeholder="Deskripsi Alat" className="w-full px-3 py-1.5 bg-white border border-slate-300 rounded" />
                          <button type="submit" disabled={actionLoading} className="w-full py-1.5 bg-purple-600 text-white rounded font-bold">Tambah Alat</button>
                        </form>
                        <div className="divide-y divide-slate-100 max-h-48 overflow-y-auto">
                          {equipments.map((eq) => (
                            <div key={eq.id} className="py-2 flex justify-between text-xs">
                              <span className="font-medium text-slate-900">{eq.name}</span>
                              <span className="text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded text-[10px] font-bold">{eq.status}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* TAMPILAN PEMINJAM (USER) */}
            {user?.role === 'user' && (
              <div className="space-y-8">
                {!membership && (
                  <div className="max-w-xl mx-auto bg-white border border-slate-200 rounded-2xl p-8 shadow-sm">
                    <h2 className="text-xl font-bold text-slate-900 mb-1 text-center">Gabung ke Organisasi / Lab</h2>
                    <p className="text-xs text-slate-500 mb-6 text-center">Masukkan Kode Redeem dari Admin dan lengkapi biodata verifikasi 2-lapis Anda.</p>
                    <form onSubmit={handleJoinOrg} className="space-y-3 text-xs">
                      <input type="text" required value={redeemCode} onChange={(e) => setRedeemCode(e.target.value.toUpperCase())} placeholder="KODE REDEEM" className="w-full px-4 py-2 border rounded-xl font-mono uppercase font-bold text-center text-sm" />
                      <input type="text" required value={fullname} onChange={(e) => setFullname(e.target.value)} placeholder="Nama Lengkap" className="w-full px-4 py-2 border rounded-xl" />
                      <input type="text" required value={identityNumber} onChange={(e) => setIdentityNumber(e.target.value)} placeholder="NIM / NIK" className="w-full px-4 py-2 border rounded-xl" />
                      <input type="tel" required value={phoneNumber} onChange={(e) => setPhoneNumber(e.target.value)} placeholder="No WhatsApp" className="w-full px-4 py-2 border rounded-xl" />
                      <textarea rows={2} required value={joinReason} onChange={(e) => setJoinReason(e.target.value)} placeholder="Alasan Keperluan Bergabung..." className="w-full px-4 py-2 border rounded-xl" />
                      <button type="submit" disabled={actionLoading} className="w-full py-2.5 bg-emerald-600 text-white rounded-xl font-bold">Kirim Permohonan</button>
                    </form>
                  </div>
                )}

                {membership && membership.status === 'PENDING' && (
                  <div className="max-w-xl mx-auto bg-amber-50 border border-amber-200 rounded-2xl p-8 text-center text-amber-900">
                    <Clock className="w-12 h-12 mx-auto text-amber-600 mb-3 animate-pulse" />
                    <h2 className="text-xl font-bold mb-1">Menunggu Persetujuan Admin Lab</h2>
                    <p className="text-xs text-amber-700">Pengajuan Anda ke {membership.organization?.name} sedang diverifikasi.</p>
                  </div>
                )}

                {membership && membership.status === 'APPROVED' && (
                  <div className="space-y-8">
                    <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
                      <h3 className="font-bold text-slate-900 text-sm mb-4">Pengajuan Peminjaman Alat Terenkripsi</h3>
                      <form onSubmit={handleCreateLoan} className="space-y-3 text-xs">
                        <select required value={selectedEqId} onChange={(e) => setSelectedEqId(e.target.value)} className="w-full px-4 py-2 border rounded-xl">
                          <option value="">-- Pilih Barang --</option>
                          {equipments.map(eq => <option key={eq.id} value={eq.id}>{eq.name} ({eq.status})</option>)}
                        </select>
                        <textarea rows={2} required value={loanPurpose} onChange={(e) => setLoanPurpose(e.target.value)} placeholder="Keperluan peminjaman..." className="w-full px-4 py-2 border rounded-xl" />
                        <div className="grid grid-cols-2 gap-3">
                          <select value={cryptoAlgo} onChange={(e) => setCryptoAlgo(e.target.value)} className="px-3 py-2 border rounded-xl font-bold text-emerald-700">
                            <option value="AES">AES-256-GCM (Modern AEAD)</option>
                            <option value="DES">DES-CBC (Legacy 56-bit)</option>
                            <option value="RC4">RC4 (Stream Cipher)</option>
                          </select>
                          <input type="file" required onChange={(e) => setDocFile(e.target.files?.[0] || null)} className="text-slate-500" />
                        </div>
                        <button type="submit" disabled={actionLoading} className="w-full py-2.5 bg-emerald-600 text-white font-bold rounded-xl">Enkripsi & Ajukan Peminjaman</button>
                      </form>
                    </div>

                    {/* Riwayat Loans */}
                    <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
                      <div className="px-6 py-3 border-b text-xs font-bold text-slate-700">Riwayat Peminjaman Anda</div>
                      <table className="w-full text-left text-xs">
                        <thead className="bg-slate-50 text-slate-500">
                          <tr><th className="py-2.5 px-6">Barang</th><th className="py-2.5 px-6">Algoritma</th><th className="py-2.5 px-6">Status</th></tr>
                        </thead>
                        <tbody className="divide-y">
                          {myLoans.map(l => (
                            <tr key={l.id} className="hover:bg-slate-50">
                              <td className="py-2.5 px-6 font-semibold">{l.equipment?.name}</td>
                              <td className="py-2.5 px-6 font-bold text-emerald-700">{l.crypto_algorithm}</td>
                              <td className="py-2.5 px-6"><span className="px-2 py-0.5 bg-amber-50 text-amber-800 rounded font-bold">{l.status}</span></td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}
              </div>
            )}
          </>
        )}

        {/* ======================================================== */}
        {/* TAB 2: BRANKAS KRIPTOGRAFI & DIGITAL SIGNATURE (VAULT)  */}
        {/* ======================================================== */}
        {activeTab === 'vault' && (
          <div className="space-y-8">
            
            {/* Banner Modul Kriptografi & Key Management */}
            <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 sm:p-8 text-white flex flex-col sm:flex-row justify-between items-start sm:items-center gap-6 shadow-xl">
              <div>
                <div className="inline-flex items-center gap-2 px-3 py-1 bg-emerald-500/10 text-emerald-400 rounded-full text-xs font-semibold mb-3 border border-emerald-500/20">
                  <Fingerprint className="w-3.5 h-3.5" /> Asymmetric RSA-2048 & SHA-256 Signature
                </div>
                <h2 className="text-2xl font-bold tracking-tight">Brankas Kriptografi Pribadi</h2>
                <p className="text-xs text-slate-400 mt-1 max-w-xl">
                  Amankan berkas privat dengan enkripsi AES-256, bubuhkan tanda tangan digital (Digital Signature), dan bagikan secara aman menggunakan pertukaran kunci RSA-OAEP.
                </p>
              </div>

              <div className="text-center sm:text-right">
                <button
                  onClick={handleGenerateRSA}
                  disabled={actionLoading}
                  className="px-5 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white font-semibold rounded-xl text-xs transition-all shadow-lg shadow-emerald-600/30 flex items-center gap-2"
                >
                  <Key className="w-4 h-4" /> Generate Pasangan Kunci RSA
                </button>
                <span className="text-[10px] text-slate-400 block mt-1.5">Kunci 2048-bit otomatis tersimpan aman</span>
              </div>
            </div>

            {/* Form Upload ke Brankas Pribadi */}
            <div className="bg-white border border-slate-200 rounded-2xl p-6 shadow-sm">
              <h3 className="font-bold text-slate-900 text-sm mb-2 flex items-center gap-2">
                <Lock className="w-4 h-4 text-emerald-600" /> Simpan Dokumen Baru ke Brankas Pribadi
              </h3>
              <p className="text-xs text-slate-500 mb-4">
                Dokumen akan dienkripsi dengan kunci AES-256 unik dan dihitung nilai integritas SHA-256-nya.
              </p>
              <form onSubmit={handleUploadVault} className="flex flex-col sm:flex-row gap-3">
                <input
                  type="file"
                  required
                  onChange={(e) => setVaultUploadFile(e.target.files?.[0] || null)}
                  className="flex-1 text-xs text-slate-500 file:mr-3 file:py-2 file:px-4 file:rounded-xl file:border-0 file:bg-emerald-50 file:text-emerald-700 file:font-semibold"
                />
                <button
                  type="submit"
                  disabled={actionLoading}
                  className="px-6 py-2.5 bg-slate-900 hover:bg-slate-800 text-white text-xs font-bold rounded-xl transition-all shadow-md"
                >
                  {actionLoading ? 'Mengenkripsi...' : 'Enkripsi & Simpan ke Brankas'}
                </button>
              </form>
            </div>

            {/* Tabel Berkas Brankas & Tanda Tangan Digital */}
            <div className="bg-white border border-slate-200 rounded-2xl shadow-sm overflow-hidden">
              <div className="px-6 py-4 border-b border-slate-200 flex justify-between items-center">
                <h3 className="font-bold text-slate-900 text-sm flex items-center gap-2">
                  <FileText className="w-4 h-4 text-emerald-600" /> Daftar Berkas Brankas & Status Tanda Tangan
                </h3>
                <span className="text-xs bg-slate-100 px-3 py-1 rounded-full font-semibold text-slate-600">
                  {vaultFiles.length} Berkas Tersimpan
                </span>
              </div>

              {vaultFiles.length === 0 ? (
                <div className="p-12 text-center text-slate-400 text-xs">
                  <FolderLock className="w-10 h-10 mx-auto text-slate-300 mb-2" />
                  Belum ada berkas di brankas digital Anda.
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider font-semibold border-b border-slate-200">
                      <tr>
                        <th className="py-3.5 px-6">Nama File</th>
                        <th className="py-3.5 px-6">Enkripsi</th>
                        <th className="py-3.5 px-6">Hash Integritas (SHA-256)</th>
                        <th className="py-3.5 px-6">Tanda Tangan Digital</th>
                        <th className="py-3.5 px-6 text-right">Aksi Kriptografi</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {vaultFiles.map((file) => (
                        <tr key={file.id} className="hover:bg-slate-50 transition-colors">
                          <td className="py-3.5 px-6 font-semibold text-slate-900">{file.filename}</td>
                          <td className="py-3.5 px-6">
                            <span className="px-2.5 py-0.5 bg-emerald-100 text-emerald-800 font-bold rounded text-[10px]">
                              {file.crypto_algorithm}-256-GCM
                            </span>
                          </td>
                          <td className="py-3.5 px-6 font-mono text-[11px] text-slate-500">
                            <span title={file.file_hash} className="cursor-help">
                              {file.file_hash.substring(0, 16)}...
                            </span>
                          </td>
                          <td className="py-3.5 px-6">
                            {file.is_signed ? (
                              <span className="inline-flex items-center gap-1 px-2.5 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 font-bold rounded-full text-[10px]">
                                <CheckCircle className="w-3 h-3" /> Signed (RSA)
                              </span>
                            ) : (
                              <span className="inline-flex items-center gap-1 px-2.5 py-0.5 bg-slate-100 text-slate-500 rounded-full text-[10px] font-medium">
                                Belum Ditandatangani
                              </span>
                            )}
                          </td>
                          <td className="py-3.5 px-6 text-right space-x-2">
                            {!file.is_signed ? (
                              <button
                                onClick={() => handleSignFile(file.id)}
                                className="px-3 py-1.5 bg-purple-50 hover:bg-purple-100 text-purple-700 font-bold rounded-lg border border-purple-200 transition-colors"
                              >
                                Tandatangani (Sign)
                              </button>
                            ) : (
                              <button
                                onClick={() => handleVerifyFile(file.id)}
                                className="px-3 py-1.5 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 font-bold rounded-lg border border-emerald-200 transition-colors"
                              >
                                Verifikasi Keaslian
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

          </div>
        )}

      </main>
    </div>
  );
}