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
  const qc = useQueryClient();
  const router = useRouter();

  const { isLoading } = useQuery({
    queryKey: ['cart'],
    queryFn: () => api.get('/cart').then((r) => { setItems(r.data.items || []); return r.data; }),
    enabled: !!user,
  });

  const removeMutation = useMutation({
    mutationFn: (productId: string) => api.delete(`/cart/${productId}`),
    onSuccess: (_, productId) => {
      removeItem(productId);
      qc.invalidateQueries({ queryKey: ['cart'] });
    },
    onError: () => toast.error('Không thể xoá sản phẩm'),
  });

  if (!user) return (
    <div className="text-center py-16">
      <p className="text-gray-500 mb-4">Vui lòng đăng nhập để xem giỏ hàng</p>
      <Link href="/auth/login" className="bg-indigo-600 text-white px-6 py-2 rounded-lg text-sm">Đăng nhập</Link>
    </div>
  );

  if (isLoading) return <div className="max-w-2xl mx-auto px-4 py-12 animate-pulse"><div className="h-32 bg-gray-200 rounded-xl" /></div>;

  const total = items.reduce((s, i) => s + i.price, 0);

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Giỏ hàng ({items.length})</h1>
      {items.length === 0 ? (
        <div className="text-center py-16">
          <p className="text-gray-500 mb-4">Giỏ hàng trống</p>
          <Link href="/products" className="text-indigo-600 hover:underline text-sm">Tiếp tục mua sắm</Link>
        </div>
      ) : (
        <>
          <div className="space-y-3">
            {items.map((item) => (
              <div key={item.product_id} className="flex items-center gap-4 bg-white rounded-xl p-4 border border-gray-100">
                <div className="relative w-16 h-12 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
                  {item.thumbnail_url && (
                    <Image src={getUploadUrl(item.thumbnail_url)} alt={item.title} fill className="object-cover" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <Link href={`/products/${item.product_id}`} className="text-sm font-medium text-gray-900 hover:text-indigo-600 line-clamp-1">
                    {item.title}
                  </Link>
                  <p className="text-xs text-gray-500">{item.seller_name}</p>
                </div>
                <PriceTag amount={item.price} className="text-sm font-bold text-indigo-600 whitespace-nowrap" />
                <button
                  onClick={() => removeMutation.mutate(item.product_id)}
                  className="text-gray-400 hover:text-red-500 p-1"
                >
                  <FiTrash2 size={16} />
                </button>
              </div>
            ))}
          </div>
          <div className="mt-6 bg-white rounded-xl p-4 border border-gray-100">
            <div className="flex justify-between text-sm mb-4">
              <span className="text-gray-600">Tổng cộng ({items.length} sản phẩm)</span>
              <PriceTag amount={total} className="font-bold text-indigo-600 text-base" />
            </div>
            <button
              onClick={() => router.push('/checkout')}
              className="w-full bg-indigo-600 text-white py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-700"
            >
              Tiến hành thanh toán
            </button>
          </div>
        </>
      )}
    </div>
  );
}
