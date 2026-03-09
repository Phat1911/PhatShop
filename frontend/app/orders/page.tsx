'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useAuthStore } from '@/lib/store';
import Link from 'next/link';
import PriceTag from '@/components/PriceTag';
import dayjs from 'dayjs';
import { useRouter } from 'next/navigation';
import clsx from 'clsx';

const STATUS_MAP: Record<string, { label: string; cls: string }> = {
  pending: { label: 'Chờ thanh toán', cls: 'bg-yellow-100 text-yellow-700' },
  paid: { label: 'Đã thanh toán', cls: 'bg-emerald-100 text-emerald-700' },
  cancelled: { label: 'Đã huỷ', cls: 'bg-gray-100 text-gray-600' },
};

export default function OrdersPage() {
  const { user, isLoading: authLoading } = useAuthStore();
  const router = useRouter();

  const { data: orders, isLoading } = useQuery({
    queryKey: ['orders'],
    queryFn: () => api.get('/orders').then((r) => r.data),
    enabled: !!user,
  });

  if (authLoading) return <div className="max-w-3xl mx-auto px-4 py-12 animate-pulse"><div className="h-64 bg-gray-200 rounded-xl" /></div>;
  if (!user) { router.push('/auth/login'); return null; }

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Đơn hàng của tôi</h1>
      {isLoading ? (
        <div className="space-y-3">{[...Array(3)].map((_, i) => <div key={i} className="h-20 bg-white rounded-xl animate-pulse" />)}</div>
      ) : orders?.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-gray-500 mb-4">Bạn chưa có đơn hàng nào</p>
          <Link href="/products" className="text-indigo-600 hover:underline text-sm">Mua sắm ngay</Link>
        </div>
      ) : (
        <div className="space-y-3">
          {orders?.map((o: { id: string; status: string; total_amount: number; created_at: string }) => {
            const s = STATUS_MAP[o.status] || STATUS_MAP.pending;
            return (
              <Link key={o.id} href={`/orders/${o.id}`}
                className="flex items-center justify-between bg-white rounded-xl p-4 border border-gray-100 hover:border-indigo-200 transition-colors">
                <div>
                  <p className="text-sm font-medium text-gray-900 font-mono">#{o.id.slice(0, 8).toUpperCase()}</p>
                  <p className="text-xs text-gray-500 mt-0.5">{dayjs(o.created_at).format('DD/MM/YYYY HH:mm')}</p>
                </div>
                <div className="flex items-center gap-4">
                  <PriceTag amount={o.total_amount} className="text-sm font-bold text-indigo-600" />
                  <span className={clsx('text-xs px-2 py-0.5 rounded-full font-medium', s.cls)}>{s.label}</span>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}
