'use client';

import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import ProductCard, { Product } from '@/components/ProductCard';
import { useSearchParams, useRouter } from 'next/navigation';
import { Suspense, useState } from 'react';

const TYPES = [
  { value: '', label: 'Tất cả' },
  { value: 'image', label: 'Hình ảnh' },
  { value: 'video', label: 'Video' },
  { value: 'website_app', label: 'Website/App' },
];

const SORTS = [
  { value: 'newest', label: 'Mới nhất' },
  { value: 'oldest', label: 'Cũ nhất' },
  { value: 'price_asc', label: 'Giá tăng dần' },
  { value: 'price_desc', label: 'Giá giảm dần' },
  { value: 'popular', label: 'Phổ biến nhất' },
];

function ProductsContent() {
  const searchParams = useSearchParams();
  const router = useRouter();

  const type = searchParams.get('type') || '';
  const sort = searchParams.get('sort') || 'newest';
  const search = searchParams.get('search') || '';
  const page = parseInt(searchParams.get('page') || '1');
  const [searchInput, setSearchInput] = useState(search);

  const { data, isLoading } = useQuery({
    queryKey: ['products', { type, sort, search, page }],
    queryFn: () => api.get('/products', { params: { type, sort, search, page, limit: 20 } }).then((r) => r.data),
  });

  const setParam = (key: string, value: string) => {
    const params = new URLSearchParams(searchParams.toString());
    if (value) params.set(key, value);
    else params.delete(key);
    params.delete('page');
    router.push('?' + params.toString());
  };

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="flex flex-col sm:flex-row gap-4 mb-6">
        <input
          type="text"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && setParam('search', searchInput)}
          placeholder="Tìm kiếm sản phẩm..."
          className="border border-gray-300 rounded-lg px-4 py-2 text-sm flex-1 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        />
        <select
          value={sort}
          onChange={(e) => setParam('sort', e.target.value)}
          className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none"
        >
          {SORTS.map((s) => <option key={s.value} value={s.value}>{s.label}</option>)}
        </select>
      </div>

      <div className="flex gap-2 mb-6 flex-wrap">
        {TYPES.map((t) => (
          <button
            key={t.value}
            onClick={() => setParam('type', t.value)}
            className={`px-4 py-1.5 rounded-full text-sm font-medium border transition-colors ${type === t.value ? 'bg-indigo-600 text-white border-indigo-600' : 'bg-white text-gray-600 border-gray-300 hover:border-indigo-400'}`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
          {[...Array(8)].map((_, i) => (
            <div key={i} className="bg-white rounded-xl aspect-[4/5] animate-pulse" />
          ))}
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
            {(data?.data || []).map((p: Product) => <ProductCard key={p.id} product={p} />)}
          </div>
          {data?.data?.length === 0 && (
            <p className="text-center text-gray-500 py-16">Không tìm thấy sản phẩm nào</p>
          )}
          {data && data.total_pages > 1 && (
            <div className="flex justify-center gap-2 mt-8">
              {[...Array(data.total_pages)].map((_, i) => (
                <button
                  key={i}
                  onClick={() => { const p = new URLSearchParams(searchParams.toString()); p.set('page', String(i + 1)); router.push('?' + p.toString()); }}
                  className={`w-8 h-8 rounded text-sm ${page === i + 1 ? 'bg-indigo-600 text-white' : 'bg-white text-gray-700 border border-gray-300'}`}
                >
                  {i + 1}
                </button>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default function ProductsPage() {
  return (
    <Suspense>
      <ProductsContent />
    </Suspense>
  );
}
