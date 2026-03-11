'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api, getUploadUrl } from '@/lib/api';
import { useParams, useRouter } from 'next/navigation';
import Image from 'next/image';
import PriceTag from '@/components/PriceTag';
import { useAuthStore, useCartStore } from '@/lib/store';
import toast from 'react-hot-toast';
import DownloadButton from '@/components/DownloadButton';
import clsx from 'clsx';

const TYPE_LABELS: Record<string, string> = {
  image: 'Hình ảnh',
  video: 'Video',
  website_app: 'Website/App',
};

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { user, isLoading: authLoading } = useAuthStore();
  const { items, setItems } = useCartStore();
  const qc = useQueryClient();
  const router = useRouter();

  const { data: product, isLoading } = useQuery({
    queryKey: ['product', id],
    queryFn: () => api.get(`/products/${id}`).then((r) => r.data),
  });

  const { data: purchaseCheck } = useQuery({
    queryKey: ['purchase-check', id],
    queryFn: () => api.get(`/downloads/check/${id}`).then((r) => r.data),
    enabled: !!user,
    staleTime: 0,
  });

  const addToCart = useMutation({
    mutationFn: () => api.post('/cart', { product_id: id }),
    onSuccess: () => {
      api.get('/cart').then((r) => setItems(r.data.items || []));
      qc.invalidateQueries({ queryKey: ['cart'] });
      toast.success('Đã thêm vào giỏ hàng');
    },
    onError: () => toast.error('Không thể thêm vào giỏ hàng'),
  });

  if (isLoading) return <div className="max-w-4xl mx-auto px-4 py-12 animate-pulse"><div className="h-64 bg-gray-200 rounded-xl" /></div>;
  if (!product) return <div className="text-center py-16 text-gray-500">Không tìm thấy sản phẩm</div>;

  const inCart = items.some((i) => i.product_id === id);
  const hasPurchased = purchaseCheck?.purchased === true;
  const thumb = product.thumbnail_url ? getUploadUrl(product.thumbnail_url) : null;
  const trailerSrc = product.trailer_url ? getUploadUrl(product.trailer_url) : null;
  // has_trailer is set by the backend when a trailer exists but the user is not signed in
  const trailerLocked = !trailerSrc && (product.has_trailer as boolean | undefined);

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
        {/* Trailer video — only shown to signed-in users */}
        {trailerSrc ? (
          <div className="relative w-full aspect-video bg-black">
            <video src={trailerSrc} controls className="w-full h-full" />
          </div>
        ) : trailerLocked ? (
          <div className="relative w-full aspect-video bg-gray-900 flex items-center justify-center overflow-hidden">
            {thumb && (
              <Image src={thumb} alt={product.title} fill className="object-cover opacity-30" />
            )}
            <div className="relative z-10 flex flex-col items-center gap-3 text-center px-4">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 text-white/80" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 10.5V6a3.75 3.75 0 10-7.5 0v4.5m11.356-1.993l1.263 12c.07.665-.45 1.243-1.119 1.243H4.25a1.125 1.125 0 01-1.12-1.243l1.264-12A1.125 1.125 0 015.513 7.5h12.974c.576 0 1.059.435 1.119 1.007z" />
              </svg>
              <p className="text-white font-medium">Đăng nhập để xem trailer</p>
              <button
                onClick={() => router.push('/auth/login')}
                className="bg-white text-gray-900 px-5 py-2 rounded-lg text-sm font-semibold hover:bg-gray-100 transition-colors"
              >
                Đăng nhập
              </button>
            </div>
          </div>
        ) : thumb ? (
          <div className="relative w-full aspect-video">
            <Image src={thumb} alt={product.title} fill className="object-cover" />
          </div>
        ) : null}
        <div className="p-6">
          <div className="flex items-start justify-between gap-4">
            <div>
              <span className="text-xs font-medium bg-indigo-100 text-indigo-700 px-2 py-0.5 rounded-full">
                {TYPE_LABELS[product.product_type] || product.product_type}
              </span>
              <h1 className="text-2xl font-bold text-gray-900 mt-2">{product.title}</h1>
              <p className="text-sm text-gray-500 mt-1">bởi {product.seller_name}</p>
            </div>
            <PriceTag amount={product.price} className="text-2xl font-bold text-indigo-600 whitespace-nowrap" />
          </div>

          {product.description && (
            <p className="text-gray-700 mt-4 leading-relaxed whitespace-pre-wrap">{product.description}</p>
          )}

          {product.tags?.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mt-4">
              {product.tags.map((tag: string) => (
                <span key={tag} className="text-xs bg-gray-100 text-gray-600 px-2 py-0.5 rounded-full">{tag}</span>
              ))}
            </div>
          )}

          <div className="flex items-center gap-4 mt-6">
            {authLoading ? (
              <div className="h-10 w-32 bg-gray-200 rounded-lg animate-pulse" />
            ) : hasPurchased ? (
              <DownloadButton productId={id} fileName={product.file_name} />
            ) : user ? (
              <button
                onClick={() => inCart ? router.push('/cart') : addToCart.mutate()}
                disabled={addToCart.isPending}
                className={clsx(
                  'px-6 py-2.5 rounded-lg text-sm font-medium transition-colors',
                  inCart
                    ? 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    : 'bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-60'
                )}
              >
                {inCart ? 'Xem giỏ hàng' : 'Thêm vào giỏ'}
              </button>
            ) : (
              <button
                onClick={() => router.push('/auth/login')}
                className="bg-indigo-600 text-white px-6 py-2.5 rounded-lg text-sm font-medium hover:bg-indigo-700"
              >
                Đăng nhập để mua
              </button>
            )}
            <div className="text-sm text-gray-500">
              {product.purchase_count} đã mua · {product.view_count} lượt xem
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
