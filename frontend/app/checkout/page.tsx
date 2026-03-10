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
  const total = items.reduce((s, i) => s + (Number(i.product?.price) || 0), 0);

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

  if (authLoading) return (
    <div className="max-w-lg mx-auto px-4 py-12">
      <div className="h-64 skeleton rounded-xl" />
    </div>
  );
  if (!user) { router.push('/auth/login'); return null; }
  if (items.length === 0) { router.push('/cart'); return null; }

  return (
    <div className="max-w-lg mx-auto px-4 py-8 animate-fade-in">
      <h1 className="text-2xl font-bold text-white mb-6">Xác nhận đơn hàng</h1>
      <div className="bg-[#16161e] rounded-2xl border border-[#1f1f2e] p-6">
        <div className="space-y-3 mb-6">
          {items.map((item) => (
            <div key={item.product_id} className="flex justify-between text-sm">
              <span className="text-gray-300 line-clamp-1 max-w-[260px]">{item.product?.title}</span>
              <PriceTag amount={item.product?.price ?? 0} className="font-medium text-[#e63946] whitespace-nowrap ml-4" />
            </div>
          ))}
        </div>
        <hr className="border-[#1f1f2e] mb-4" />
        <div className="flex justify-between mb-6">
          <span className="font-semibold text-gray-100">Tổng cộng</span>
          <PriceTag amount={total} className="font-bold text-white text-lg" />
        </div>
        <button
          onClick={() => createOrder.mutate()}
          disabled={createOrder.isPending}
          className="w-full bg-[#e63946] hover:bg-[#ff4d5a] text-white py-3 rounded-xl text-sm font-semibold transition-colors disabled:opacity-60 animate-pulse-glow"
        >
          {createOrder.isPending ? 'Đang tạo đơn...' : 'Tạo đơn & Thanh toán'}
        </button>
        <p className="text-xs text-center text-gray-500 mt-3">
          Bạn sẽ được chuyển đến trang thanh toán chuyển khoản ngân hàng
        </p>
      </div>
    </div>
  );
}
