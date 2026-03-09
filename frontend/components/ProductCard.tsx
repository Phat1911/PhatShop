import Link from 'next/link';
import Image from 'next/image';
import PriceTag from './PriceTag';
import { getUploadUrl } from '@/lib/api';
import clsx from 'clsx';

export interface Product {
  id: string;
  title: string;
  slug: string;
  description: string;
  product_type: string;
  price: number;
  thumbnail_url: string;
  preview_urls: string[];
  tags: string[];
  seller_name: string;
  category_name: string;
  purchase_count: number;
  view_count: number;
  is_published: boolean;
  created_at: string;
}

const TYPE_LABELS: Record<string, string> = {
  image: 'Hình ảnh',
  video: 'Video',
  website_app: 'Website/App',
};

const TYPE_COLORS: Record<string, string> = {
  image: 'bg-emerald-100 text-emerald-700',
  video: 'bg-pink-100 text-pink-700',
  website_app: 'bg-violet-100 text-violet-700',
};

export default function ProductCard({ product }: { product: Product }) {
  const thumb = product.thumbnail_url ? getUploadUrl(product.thumbnail_url) : null;

  return (
    <Link href={`/products/${product.id}`} className="group block bg-white rounded-xl shadow-sm hover:shadow-md transition-shadow border border-gray-100 overflow-hidden">
      <div className="relative aspect-video bg-gray-100">
        {thumb ? (
          <Image src={thumb} alt={product.title} fill className="object-cover group-hover:scale-105 transition-transform" />
        ) : (
          <div className="absolute inset-0 flex items-center justify-center text-gray-300 text-4xl">
            {product.product_type === 'video' ? '🎬' : product.product_type === 'image' ? '🖼️' : '💻'}
          </div>
        )}
        <span className={clsx('absolute top-2 left-2 text-xs font-medium px-2 py-0.5 rounded-full', TYPE_COLORS[product.product_type])}>
          {TYPE_LABELS[product.product_type] || product.product_type}
        </span>
      </div>
      <div className="p-3">
        <p className="text-sm font-semibold text-gray-900 line-clamp-2 leading-snug">{product.title}</p>
        <p className="text-xs text-gray-500 mt-1">{product.seller_name}</p>
        <div className="mt-2 flex items-center justify-between">
          <PriceTag amount={product.price} className="text-indigo-600 font-bold text-sm" />
          <span className="text-xs text-gray-400">{product.purchase_count} đã bán</span>
        </div>
      </div>
    </Link>
  );
}
