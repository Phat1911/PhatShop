'use client';

import { api, UPLOADS_URL } from '@/lib/api';
import { useState } from 'react';
import { FiDownload, FiLoader } from 'react-icons/fi';
import toast from 'react-hot-toast';

export default function DownloadButton({ productId, fileName }: { productId: string; fileName?: string }) {
  const [loading, setLoading] = useState(false);

  const handleDownload = async () => {
    setLoading(true);
    try {
      const { data } = await api.get(`/downloads/request/${productId}`);
      window.location.href = `${UPLOADS_URL.replace('/uploads', '')}/api/v1/downloads/file?token=${data.token}`;
    } catch {
      toast.error('Không thể tạo link tải xuống');
    } finally {
      setLoading(false);
    }
  };

  return (
    <button
      onClick={handleDownload}
      disabled={loading}
      className="flex items-center gap-2 bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 disabled:opacity-60 text-sm font-medium"
    >
      {loading ? <FiLoader className="animate-spin" size={16} /> : <FiDownload size={16} />}
      {loading ? 'Đang tạo link...' : `Tải xuống${fileName ? ` (${fileName})` : ''}`}
    </button>
  );
}
