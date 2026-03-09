'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { api } from '@/lib/api';
import { useAuthStore } from '@/lib/store';
import toast from 'react-hot-toast';

export default function RegisterPage() {
  const [form, setForm] = useState({ username: '', email: '', password: '', display_name: '' });
  const [loading, setLoading] = useState(false);
  const setAuth = useAuthStore((s) => s.setAuth);
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const { data } = await api.post('/auth/register', form);
      setAuth(data.user, data.token);
      toast.success('Đăng ký thành công!');
      router.push('/products');
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Đăng ký thất bại';
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm bg-white rounded-2xl shadow-sm border border-gray-100 p-8">
        <h1 className="text-2xl font-bold text-gray-900 mb-1">Đăng ký</h1>
        <p className="text-sm text-gray-500 mb-6">Tạo tài khoản PhatShop miễn phí</p>
        <form onSubmit={handleSubmit} className="space-y-4">
          {[
            { label: 'Tên hiển thị', key: 'display_name', type: 'text', placeholder: 'Nguyễn Văn A' },
            { label: 'Tên đăng nhập', key: 'username', type: 'text', placeholder: 'nguyen_van_a' },
            { label: 'Email', key: 'email', type: 'email', placeholder: 'ban@email.com' },
            { label: 'Mật khẩu', key: 'password', type: 'password', placeholder: '••••••••' },
          ].map(({ label, key, type, placeholder }) => (
            <div key={key}>
              <label className="block text-sm font-medium text-gray-700 mb-1">{label}</label>
              <input
                type={type}
                value={form[key as keyof typeof form]}
                onChange={(e) => setForm({ ...form, [key]: e.target.value })}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder={placeholder}
                required
              />
            </div>
          ))}
          <button
            type="submit" disabled={loading}
            className="w-full bg-indigo-600 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-700 disabled:opacity-60"
          >
            {loading ? 'Đang đăng ký...' : 'Đăng ký'}
          </button>
        </form>
        <p className="text-sm text-center text-gray-500 mt-4">
          Đã có tài khoản?{' '}
          <Link href="/auth/login" className="text-indigo-600 hover:underline">Đăng nhập</Link>
        </p>
      </div>
    </div>
  );
}
