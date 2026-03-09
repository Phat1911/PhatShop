'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import Link from 'next/link';
import PriceTag from '@/components/PriceTag';
import { FiPlus, FiTrash2, FiEye, FiEyeOff } from 'react-icons/fi';
import toast from 'react-hot-toast';
import clsx from 'clsx';

interface AdminProduct {
  id: string;
  title: string;
  product_type: string;
  price: number;
  is_published: boolean;
  purchase_count: number;
  seller_name: string;
}

const TYPE_LABELS: Record<string, string> = {
  image: 'Hình ảnh', video: 'Video', website_app: 'Website/App',
};

export default function AdminProductsPage() {
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['admin-products'],
    queryFn: () => api.get('/admin/products', { params: { limit: 50 } }).then((r) => r.data),
  });

  const togglePublish = useMutation({
    mutationFn: ({ id, published }: { id: string; published: boolean }) =>
      api.patch(`/admin/products/${id}/publish`, { is_published: published }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin-products'] }); toast.success('Đã cập nhật'); },
    onError: () => toast.error('Cập nhật thất bại'),
  });

  const deleteProd = useMutation({
    mutationFn: (id: string) => api.delete(`/admin/products/${id}`),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['admin-products'] }); toast.success('Đã xoá sản phẩm'); },
    onError: () => toast.error('Xoá thất bại'),
  });

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Sản phẩm</h1>
        <Link href="/admin/products/new" className="flex items-center gap-2 bg-indigo-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-indigo-700">
          <FiPlus size={16} /> Thêm sản phẩm
        </Link>
      </div>
      {isLoading ? (
        <div className="space-y-2">{[...Array(5)].map((_, i) => <div key={i} className="h-12 bg-gray-100 rounded-lg animate-pulse" />)}</div>
      ) : (
        <div className="bg-white rounded-xl border border-gray-100 overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-100">
              <tr>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Tên sản phẩm</th>
                <th className="text-left px-4 py-3 font-medium text-gray-600">Loại</th>
                <th className="text-right px-4 py-3 font-medium text-gray-600">Giá</th>
                <th className="text-center px-4 py-3 font-medium text-gray-600">Đã bán</th>
                <th className="text-center px-4 py-3 font-medium text-gray-600">Trạng thái</th>
                <th className="text-center px-4 py-3 font-medium text-gray-600">Hành động</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              {(data?.data || []).map((p: AdminProduct) => (
                <tr key={p.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 font-medium text-gray-900 max-w-[200px] truncate">{p.title}</td>
                  <td className="px-4 py-3 text-gray-600">{TYPE_LABELS[p.product_type] || p.product_type}</td>
                  <td className="px-4 py-3 text-right"><PriceTag amount={p.price} className="text-indigo-600 font-medium" /></td>
                  <td className="px-4 py-3 text-center text-gray-600">{p.purchase_count}</td>
                  <td className="px-4 py-3 text-center">
                    <span className={clsx('text-xs px-2 py-0.5 rounded-full font-medium', p.is_published ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-600')}>
                      {p.is_published ? 'Đang bán' : 'Ẩn'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-center gap-2">
                      <button
                        onClick={() => togglePublish.mutate({ id: p.id, published: !p.is_published })}
                        className="p-1.5 text-gray-500 hover:text-indigo-600 rounded"
                        title={p.is_published ? 'Ẩn' : 'Hiện'}
                      >
                        {p.is_published ? <FiEyeOff size={16} /> : <FiEye size={16} />}
                      </button>
                      <button
                        onClick={() => { if (confirm('Xoá sản phẩm này?')) deleteProd.mutate(p.id); }}
                        className="p-1.5 text-gray-500 hover:text-red-600 rounded"
                      >
                        <FiTrash2 size={16} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
