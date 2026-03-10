import type { Metadata } from 'next';
import './globals.css';
import Providers from './providers';
import Navbar from '@/components/Navbar';

export const metadata: Metadata = {
  title: 'PhatShop — Chợ Tài Nguyên Số',
  description: 'Mua bán hình ảnh, video và website/app chất lượng cao',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="vi">
      <body className="min-h-screen bg-[#0a0a0f]">
        <Providers>
          <Navbar />
          <main>{children}</main>
        </Providers>
      </body>
    </html>
  );
}
