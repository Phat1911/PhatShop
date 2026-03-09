'use client';

import { useMutation } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useCartStore, useAuthStore } from '@/lib/store';
import PriceTag from '@/components/PriceTag';
import { useRouter } from 'next/navigation';
import toast from 'react-hot-toast';

export default function CheckoutPage() {
  const { items } = useCartStore();
  const { user, isLoading: authLoading } = useAuthStore();
  const router = useRouter();
  const total = items.reduce((s, i) => s + i.price, 0);

  const createOrder = useMutation({
    mutationFn: () => api.post('/orders').then((r) => r.data),
    onSuccess: (order) => {
      router.push(`/checkout/payment/${order.id}`);
    },
    onError: (err: unknown) => {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Tạo đơn hàng thất bại';
      toast.error(msg);
    },
  });

  if (authLoading) return <div className="max-w-lg mx-auto px-4 py-12 animate-pulse"><div className="h-64 bg-gray-200 rounded-xl" /></div>;
  if (!user) { router.push('/auth/login'); return null; }
  if (items.length === 0) { router.push('/cart'); return null; }

  return (
    <div className="max-w-lg mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Xác nhận đơn hàng</h1>
      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
        <div className="space-y-3 mb-6">
          {items.map((item) => (
            <div key={item.product_id} className="flex justify-between text-sm">
              <span className="text-gray-700 line-clamp-1 max-w-[260px]">{item.title}</span>
              <PriceTag amount={item.price} className="font-medium text-gray-900 whitespace-nowrap ml-4" />
            </div>
          ))}
        </div>
        <hr className="border-gray-100 mb-4" />
        <div className="flex justify-between mb-6">
          <span className="font-semibold text-gray-900">Tổng cộng</span>
          <PriceTag amount={total} className="font-bold text-indigo-600 text-lg" />
        </div>
        <button
          onClick={() => createOrder.mutate()}
          disabled={createOrder.isPending}
          className="w-full bg-indigo-600 text-white py-3 rounded-xl text-sm font-semibold hover:bg-indigo-700 disabled:opacity-60"
        >
          {createOrder.isPending ? 'Đang tạo đơn...' : 'Tạo đơn & Thanh toán'}
        </button>
        <p className="text-xs text-center text-gray-400 mt-3">
          Bạn sẽ được chuyển đến trang thanh toán chuyển khoản ngân hàng
        </p>
      </div>
    </div>
  );
}
