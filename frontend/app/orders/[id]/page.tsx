'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useAuthStore } from '@/lib/store';
import { useParams, useRouter } from 'next/navigation';
import PriceTag from '@/components/PriceTag';
import DownloadButton from '@/components/DownloadButton';
import dayjs from 'dayjs';
import clsx from 'clsx';
import Link from 'next/link';
import { FiArrowLeft, FiClock } from 'react-icons/fi';
import { useEffect, useRef } from 'react';
import toast from 'react-hot-toast';

const STATUS_MAP: Record<string, { label: string; cls: string }> = {
  pending: { label: 'Chờ thanh toán', cls: 'bg-yellow-100 text-yellow-700' },
  paid: { label: 'Đã thanh toán', cls: 'bg-emerald-100 text-emerald-700' },
  cancelled: { label: 'Đã huỷ', cls: 'bg-gray-100 text-gray-600' },
};

interface OrderItem {
  product_id: string;
  price: number;
  product?: { title: string; file_name: string };
}

export default function OrderDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { user, isLoading: authLoading } = useAuthStore();
  const router = useRouter();
  const prevStatus = useRef<string | null>(null);

  const { data: order, isLoading } = useQuery({
    queryKey: ['order', id],
    queryFn: () => api.get(`/orders/${id}`).then((r) => r.data),
    enabled: !!user,
    refetchInterval: (query) => {
      if (query.state.data?.status === 'paid') return false;
      if (query.state.data?.status === 'pending') return 5000;
      return false;
    },
  });

  useEffect(() => {
    if (!order) return;
    if (prevStatus.current === 'pending' && order.status === 'paid') {
      toast.success('Thanh toán xác nhận! Bạn có thể tải xuống ngay bây giờ.');
    }
    prevStatus.current = order.status;
  }, [order?.status]);

  if (authLoading) return <div className="max-w-2xl mx-auto px-4 py-12 animate-pulse"><div className="h-64 bg-gray-200 rounded-xl" /></div>;
  if (!user) { router.push('/auth/login'); return null; }
  if (isLoading) return <div className="max-w-2xl mx-auto px-4 py-12 animate-pulse"><div className="h-64 bg-gray-200 rounded-xl" /></div>;
  if (!order) return <div className="text-center py-16 text-gray-500">Không tìm thấy đơn hàng</div>;

  const s = STATUS_MAP[order.status] || STATUS_MAP.pending;

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <Link href="/orders" className="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-900 mb-6">
        <FiArrowLeft size={16} /> Quay lại đơn hàng
      </Link>
      <div className="bg-white rounded-2xl border border-gray-100 p-6">
        <div className="flex items-start justify-between mb-6">
          <div>
            <h1 className="text-lg font-bold text-gray-900 font-mono">#{order.id.slice(0, 8).toUpperCase()}</h1>
            <p className="text-xs text-gray-500 mt-0.5">{dayjs(order.created_at).format('DD/MM/YYYY HH:mm')}</p>
          </div>
          <span className={clsx('text-sm px-3 py-1 rounded-full font-medium', s.cls)}>{s.label}</span>
        </div>

        {order.status === 'pending' && (
          <div className="flex items-center gap-2 bg-yellow-50 border border-yellow-200 rounded-xl px-4 py-3 mb-4">
            <FiClock className="text-yellow-500 animate-pulse flex-shrink-0" size={16} />
            <p className="text-sm text-yellow-700">Đang chờ xác nhận thanh toán. Trang sẽ tự động cập nhật.</p>
          </div>
        )}

        <h2 className="text-sm font-semibold text-gray-700 mb-3">Sản phẩm</h2>
        <div className="space-y-3">
          {order.items?.map((item: OrderItem) => (
            <div key={item.product_id} className="flex items-center justify-between p-3 bg-gray-50 rounded-xl">
              <div className="min-w-0">
                <p className="text-sm font-medium text-gray-900 line-clamp-1">{item.product?.title}</p>
                <PriceTag amount={item.price} className="text-xs text-indigo-600 font-medium" />
              </div>
              {order.status === 'paid' && (
                <div className="ml-4 flex-shrink-0">
                  <DownloadButton productId={item.product_id} fileName={item.product?.file_name ?? ''} />
                </div>
              )}
            </div>
          ))}
        </div>

        <hr className="my-4 border-gray-100" />
        <div className="flex justify-between">
          <span className="font-semibold text-gray-900">Tổng cộng</span>
          <PriceTag amount={order.total_amount} className="font-bold text-indigo-600 text-lg" />
        </div>
      </div>
    </div>
  );
}
