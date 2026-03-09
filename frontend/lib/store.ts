import { create } from 'zustand';
import Cookies from 'js-cookie';

export interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url: string;
  role: string;
  created_at: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  setAuth: (user: User, token: string) => void;
  clearAuth: () => void;
  setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: null,
  isLoading: true,
  setAuth: (user, token) => {
    Cookies.set('phatshop_token', token, { expires: 7 });
    set({ user, token, isLoading: false });
  },
  clearAuth: () => {
    Cookies.remove('phatshop_token');
    set({ user: null, token: null, isLoading: false });
  },
  setLoading: (loading) => set({ isLoading: loading }),
}));

export interface CartItem {
  id: string;
  product_id: string;
  title: string;
  thumbnail_url: string;
  product_type: string;
  price: number;
  seller_name: string;
}

interface CartState {
  items: CartItem[];
  setItems: (items: CartItem[]) => void;
  removeItem: (productId: string) => void;
  clearCart: () => void;
}

export const useCartStore = create<CartState>((set) => ({
  items: [],
  setItems: (items) => set({ items }),
  removeItem: (productId) =>
    set((state) => ({ items: state.items.filter((i) => i.product_id !== productId) })),
  clearCart: () => set({ items: [] }),
}));
