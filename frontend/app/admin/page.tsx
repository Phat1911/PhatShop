'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';

export default function AdminDashboard() {
  const { data: stats } = useQuery({
    queryKey: ['admin-stats'],
    queryFn: () => api.get('/admin/stats').then((r) => r.data),
  });

  const cards = [
    { label: 'Tổng sản phẩm', value: stats?.total_products ?? '—' },
    { label: 'Tổng đơn hàng', value: stats?.total_orders ?? '—' },
    { label: 'Tổng doanh thu', value: stats?.total_revenue ? new Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }).format(stats.total_revenue) : '—' },
    { label: 'Tổng người dùng', value: stats?.total_users ?? '—' },
  ];

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Tổng quan</h1>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        {cards.map((c) => (
          <div key={c.label} className="bg-white rounded-xl border border-gray-100 p-5">
            <p className="text-sm text-gray-500">{c.label}</p>
            <p className="text-2xl font-bold text-gray-900 mt-1">{c.value}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
