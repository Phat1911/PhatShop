'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api, getUploadUrl } from '@/lib/api';
import { useAuthStore, useCartStore } from '@/lib/store';
import Link from 'next/link';
import Image from 'next/image';
import PriceTag from '@/components/PriceTag';
import { FiTrash2 } from 'react-icons/fi';
import toast from 'react-hot-toast';
import { useRouter } from 'next/navigation';

export default function CartPage() {
  const { user } = useAuthStore();
  const { items, setItems, removeItem } = useCartStore();
  const queryClient = useQueryClient();
  const router = useRouter();

  useQuery({
    queryKey: ['cart'],
    queryFn: async () => {
      const r = await api.get('/cart');
      setItems(r.data.items || []);
      return r.data;
    },
    enabled: !!user,
  });

  const total = items.reduce((s, i) => s + (Number(i.product?.price) || 0), 0);

  const removeMutation = useMutation({
    mutationFn: (productId: string) => api.delete(`/cart/${productId}`),
    onSuccess: (_, productId) => {
      removeItem(productId);
      queryClient.invalidateQueries({ queryKey: ['cart'] });
      toast.success('Đã xoá khỏi giỏ hàng');
    },
    onError: () => toast.error('Không thể xoá sản phẩm'),
  });

  if (!user) {
    return (
      <div className="max-w-xl mx-auto px-4 py-20 text-center animate-fade-in">
        <p className="text-gray-400 mb-4">Vui lòng đăng nhập để xem giỏ hàng</p>
        <Link href="/auth/login" className="bg-[#e63946] hover:bg-[#ff4d5a] text-white px-6 py-2.5 rounded-lg text-sm font-medium transition-colors">
          Đăng nhập
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto px-4 py-8 animate-fade-in">
      <h1 className="text-2xl font-bold text-white mb-6">Giỏ hàng</h1>

      {items.length === 0 ? (
        <div className="text-center py-20">
          <p className="text-gray-500 mb-4">Giỏ hàng trống</p>
          <Link href="/products" className="text-[#2563eb] hover:text-[#3b82f6] text-sm transition-colors">
            Khám phá sản phẩm
          </Link>
        </div>
      ) : (
        <>
          <div className="space-y-3 mb-6">
            {items.map((item) => (
              <div key={item.id} className="flex items-center gap-4 bg-[#16161e] border border-[#1f1f2e] rounded-xl p-3 card-hover">
                <div className="relative w-20 h-14 rounded-lg overflow-hidden flex-shrink-0 bg-[#111118]">
                  {item.product?.thumbnail_url && (
                    <Image
                      src={getUploadUrl(item.product.thumbnail_url)}
                      alt={item.product?.title || ''}
                      fill
                      className="object-cover"
                    />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-100 truncate">{item.product?.title}</p>
                  <p className="text-xs text-gray-500 mt-0.5 capitalize">{item.product?.product_type}</p>
                </div>
                <PriceTag amount={item.product?.price ?? 0} className="text-sm font-bold text-[#e63946] whitespace-nowrap" />
                <button
                  onClick={() => removeMutation.mutate(item.product_id)}
                  disabled={removeMutation.isPending}
                  className="text-gray-600 hover:text-[#e63946] transition-colors flex-shrink-0"
                >
                  <FiTrash2 size={18} />
                </button>
              </div>
            ))}
          </div>

          <div className="bg-[#16161e] border border-[#1f1f2e] rounded-xl p-4">
            <div className="flex items-center justify-between mb-4">
              <span className="text-gray-400 text-sm">Tổng cộng ({items.length} sản phẩm)</span>
              <PriceTag amount={total} className="font-bold text-white text-base" />
            </div>
            <button
              onClick={() => router.push('/checkout')}
              className="w-full bg-[#e63946] hover:bg-[#ff4d5a] text-white py-2.5 rounded-lg text-sm font-medium transition-colors animate-pulse-glow"
            >
              Tiến hành thanh toán
            </button>
          </div>
        </>
      )}
    </div>
  );
}
